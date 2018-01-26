package sarama

import (
	"encoding/binary"
	"time"
)

const (
	controlMask           = 0x20
	maximumRecordOverhead = 5*binary.MaxVarintLen32 + binary.MaxVarintLen64 + 1
)

type RecordHeader struct {
	Key   []byte
	Value []byte
}

func (h *RecordHeader) encode(pe packetEncoder) error {
	if err := pe.putVarintBytes(h.Key); err != nil {
		return err
	}
	return pe.putVarintBytes(h.Value)
}

func (h *RecordHeader) decode(pd packetDecoder) (err error) {
	if h.Key, err = pd.getVarintBytes(); err != nil {
		return err
	}

	if h.Value, err = pd.getVarintBytes(); err != nil {
		return err
	}
	return nil
}

type Record struct {
	Attributes     int8
	TimestampDelta time.Duration
	OffsetDelta    int64
	Key            []byte
	Value          []byte
	Headers        []*RecordHeader

	length varintLengthField
}

func (r *Record) encode(pe packetEncoder) error {
	pe.push(&r.length)
	pe.putInt8(r.Attributes)
	pe.putVarint(int64(r.TimestampDelta / time.Millisecond))
	pe.putVarint(r.OffsetDelta)
	if err := pe.putVarintBytes(r.Key); err != nil {
		return err
	}
	if err := pe.putVarintBytes(r.Value); err != nil {
		return err
	}
	pe.putVarint(int64(len(r.Headers)))

	for _, h := range r.Headers {
		if err := h.encode(pe); err != nil {
			return err
		}
	}

	return pe.pop()
}

func (r *Record) decode(pd packetDecoder) (err error) {
	if err = pd.push(&r.length); err != nil {
		return err
	}

	if r.Attributes, err = pd.getInt8(); err != nil {
		return err
	}

	timestamp, err := pd.getVarint()
	if err != nil {
		return err
	}
	r.TimestampDelta = time.Duration(timestamp) * time.Millisecond

	if r.OffsetDelta, err = pd.getVarint(); err != nil {
		return err
	}

	if r.Key, err = pd.getVarintBytes(); err != nil {
		return err
	}

	if r.Value, err = pd.getVarintBytes(); err != nil {
		return err
	}

	numHeaders, err := pd.getVarint()
	if err != nil {
		return err
	}

	if numHeaders >= 0 {
		r.Headers = make([]*RecordHeader, numHeaders)
	}
	for i := int64(0); i < numHeaders; i++ {
		hdr := new(RecordHeader)
		if err := hdr.decode(pd); err != nil {
			return err
		}
		r.Headers[i] = hdr
	}

	return pd.pop()
}
