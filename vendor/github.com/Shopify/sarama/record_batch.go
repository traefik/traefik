package sarama

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/eapache/go-xerial-snappy"
	"github.com/pierrec/lz4"
)

const recordBatchOverhead = 49

type recordsArray []*Record

func (e recordsArray) encode(pe packetEncoder) error {
	for _, r := range e {
		if err := r.encode(pe); err != nil {
			return err
		}
	}
	return nil
}

func (e recordsArray) decode(pd packetDecoder) error {
	for i := range e {
		rec := &Record{}
		if err := rec.decode(pd); err != nil {
			return err
		}
		e[i] = rec
	}
	return nil
}

type RecordBatch struct {
	FirstOffset           int64
	PartitionLeaderEpoch  int32
	Version               int8
	Codec                 CompressionCodec
	Control               bool
	LastOffsetDelta       int32
	FirstTimestamp        time.Time
	MaxTimestamp          time.Time
	ProducerID            int64
	ProducerEpoch         int16
	FirstSequence         int32
	Records               []*Record
	PartialTrailingRecord bool

	compressedRecords []byte
	recordsLen        int // uncompressed records size
}

func (b *RecordBatch) encode(pe packetEncoder) error {
	if b.Version != 2 {
		return PacketEncodingError{fmt.Sprintf("unsupported compression codec (%d)", b.Codec)}
	}
	pe.putInt64(b.FirstOffset)
	pe.push(&lengthField{})
	pe.putInt32(b.PartitionLeaderEpoch)
	pe.putInt8(b.Version)
	pe.push(newCRC32Field(crcCastagnoli))
	pe.putInt16(b.computeAttributes())
	pe.putInt32(b.LastOffsetDelta)

	if err := (Timestamp{&b.FirstTimestamp}).encode(pe); err != nil {
		return err
	}

	if err := (Timestamp{&b.MaxTimestamp}).encode(pe); err != nil {
		return err
	}

	pe.putInt64(b.ProducerID)
	pe.putInt16(b.ProducerEpoch)
	pe.putInt32(b.FirstSequence)

	if err := pe.putArrayLength(len(b.Records)); err != nil {
		return err
	}

	if b.compressedRecords == nil {
		if err := b.encodeRecords(pe); err != nil {
			return err
		}
	}
	if err := pe.putRawBytes(b.compressedRecords); err != nil {
		return err
	}

	if err := pe.pop(); err != nil {
		return err
	}
	return pe.pop()
}

func (b *RecordBatch) decode(pd packetDecoder) (err error) {
	if b.FirstOffset, err = pd.getInt64(); err != nil {
		return err
	}

	batchLen, err := pd.getInt32()
	if err != nil {
		return err
	}

	if b.PartitionLeaderEpoch, err = pd.getInt32(); err != nil {
		return err
	}

	if b.Version, err = pd.getInt8(); err != nil {
		return err
	}

	if err = pd.push(&crc32Field{polynomial: crcCastagnoli}); err != nil {
		return err
	}

	attributes, err := pd.getInt16()
	if err != nil {
		return err
	}
	b.Codec = CompressionCodec(int8(attributes) & compressionCodecMask)
	b.Control = attributes&controlMask == controlMask

	if b.LastOffsetDelta, err = pd.getInt32(); err != nil {
		return err
	}

	if err = (Timestamp{&b.FirstTimestamp}).decode(pd); err != nil {
		return err
	}

	if err = (Timestamp{&b.MaxTimestamp}).decode(pd); err != nil {
		return err
	}

	if b.ProducerID, err = pd.getInt64(); err != nil {
		return err
	}

	if b.ProducerEpoch, err = pd.getInt16(); err != nil {
		return err
	}

	if b.FirstSequence, err = pd.getInt32(); err != nil {
		return err
	}

	numRecs, err := pd.getArrayLength()
	if err != nil {
		return err
	}
	if numRecs >= 0 {
		b.Records = make([]*Record, numRecs)
	}

	bufSize := int(batchLen) - recordBatchOverhead
	recBuffer, err := pd.getRawBytes(bufSize)
	if err != nil {
		if err == ErrInsufficientData {
			b.PartialTrailingRecord = true
			b.Records = nil
			return nil
		}
		return err
	}

	if err = pd.pop(); err != nil {
		return err
	}

	switch b.Codec {
	case CompressionNone:
	case CompressionGZIP:
		reader, err := gzip.NewReader(bytes.NewReader(recBuffer))
		if err != nil {
			return err
		}
		if recBuffer, err = ioutil.ReadAll(reader); err != nil {
			return err
		}
	case CompressionSnappy:
		if recBuffer, err = snappy.Decode(recBuffer); err != nil {
			return err
		}
	case CompressionLZ4:
		reader := lz4.NewReader(bytes.NewReader(recBuffer))
		if recBuffer, err = ioutil.ReadAll(reader); err != nil {
			return err
		}
	default:
		return PacketDecodingError{fmt.Sprintf("invalid compression specified (%d)", b.Codec)}
	}

	b.recordsLen = len(recBuffer)
	err = decode(recBuffer, recordsArray(b.Records))
	if err == ErrInsufficientData {
		b.PartialTrailingRecord = true
		b.Records = nil
		return nil
	}
	return err
}

func (b *RecordBatch) encodeRecords(pe packetEncoder) error {
	var raw []byte
	if b.Codec != CompressionNone {
		var err error
		if raw, err = encode(recordsArray(b.Records), nil); err != nil {
			return err
		}
		b.recordsLen = len(raw)
	}

	switch b.Codec {
	case CompressionNone:
		offset := pe.offset()
		if err := recordsArray(b.Records).encode(pe); err != nil {
			return err
		}
		b.recordsLen = pe.offset() - offset
	case CompressionGZIP:
		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		if _, err := writer.Write(raw); err != nil {
			return err
		}
		if err := writer.Close(); err != nil {
			return err
		}
		b.compressedRecords = buf.Bytes()
	case CompressionSnappy:
		b.compressedRecords = snappy.Encode(raw)
	case CompressionLZ4:
		var buf bytes.Buffer
		writer := lz4.NewWriter(&buf)
		if _, err := writer.Write(raw); err != nil {
			return err
		}
		if err := writer.Close(); err != nil {
			return err
		}
		b.compressedRecords = buf.Bytes()
	default:
		return PacketEncodingError{fmt.Sprintf("unsupported compression codec (%d)", b.Codec)}
	}

	return nil
}

func (b *RecordBatch) computeAttributes() int16 {
	attr := int16(b.Codec) & int16(compressionCodecMask)
	if b.Control {
		attr |= controlMask
	}
	return attr
}

func (b *RecordBatch) addRecord(r *Record) {
	b.Records = append(b.Records, r)
}
