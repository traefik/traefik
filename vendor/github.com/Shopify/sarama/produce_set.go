package sarama

import (
	"encoding/binary"
	"time"
)

type partitionSet struct {
	msgs          []*ProducerMessage
	recordsToSend Records
	bufferBytes   int
}

type produceSet struct {
	parent *asyncProducer
	msgs   map[string]map[int32]*partitionSet

	bufferBytes int
	bufferCount int
}

func newProduceSet(parent *asyncProducer) *produceSet {
	return &produceSet{
		msgs:   make(map[string]map[int32]*partitionSet),
		parent: parent,
	}
}

func (ps *produceSet) add(msg *ProducerMessage) error {
	var err error
	var key, val []byte

	if msg.Key != nil {
		if key, err = msg.Key.Encode(); err != nil {
			return err
		}
	}

	if msg.Value != nil {
		if val, err = msg.Value.Encode(); err != nil {
			return err
		}
	}

	timestamp := msg.Timestamp
	if msg.Timestamp.IsZero() {
		timestamp = time.Now()
	}

	partitions := ps.msgs[msg.Topic]
	if partitions == nil {
		partitions = make(map[int32]*partitionSet)
		ps.msgs[msg.Topic] = partitions
	}

	var size int

	set := partitions[msg.Partition]
	if set == nil {
		if ps.parent.conf.Version.IsAtLeast(V0_11_0_0) {
			batch := &RecordBatch{
				FirstTimestamp:   timestamp,
				Version:          2,
				ProducerID:       -1, /* No producer id */
				Codec:            ps.parent.conf.Producer.Compression,
				CompressionLevel: ps.parent.conf.Producer.CompressionLevel,
			}
			set = &partitionSet{recordsToSend: newDefaultRecords(batch)}
			size = recordBatchOverhead
		} else {
			set = &partitionSet{recordsToSend: newLegacyRecords(new(MessageSet))}
		}
		partitions[msg.Partition] = set
	}

	set.msgs = append(set.msgs, msg)
	if ps.parent.conf.Version.IsAtLeast(V0_11_0_0) {
		// We are being conservative here to avoid having to prep encode the record
		size += maximumRecordOverhead
		rec := &Record{
			Key:            key,
			Value:          val,
			TimestampDelta: timestamp.Sub(set.recordsToSend.RecordBatch.FirstTimestamp),
		}
		size += len(key) + len(val)
		if len(msg.Headers) > 0 {
			rec.Headers = make([]*RecordHeader, len(msg.Headers))
			for i := range msg.Headers {
				rec.Headers[i] = &msg.Headers[i]
				size += len(rec.Headers[i].Key) + len(rec.Headers[i].Value) + 2*binary.MaxVarintLen32
			}
		}
		set.recordsToSend.RecordBatch.addRecord(rec)
	} else {
		msgToSend := &Message{Codec: CompressionNone, Key: key, Value: val}
		if ps.parent.conf.Version.IsAtLeast(V0_10_0_0) {
			msgToSend.Timestamp = timestamp
			msgToSend.Version = 1
		}
		set.recordsToSend.MsgSet.addMessage(msgToSend)
		size = producerMessageOverhead + len(key) + len(val)
	}

	set.bufferBytes += size
	ps.bufferBytes += size
	ps.bufferCount++

	return nil
}

func (ps *produceSet) buildRequest() *ProduceRequest {
	req := &ProduceRequest{
		RequiredAcks: ps.parent.conf.Producer.RequiredAcks,
		Timeout:      int32(ps.parent.conf.Producer.Timeout / time.Millisecond),
	}
	if ps.parent.conf.Version.IsAtLeast(V0_10_0_0) {
		req.Version = 2
	}
	if ps.parent.conf.Version.IsAtLeast(V0_11_0_0) {
		req.Version = 3
	}

	for topic, partitionSet := range ps.msgs {
		for partition, set := range partitionSet {
			if req.Version >= 3 {
				// If the API version we're hitting is 3 or greater, we need to calculate
				// offsets for each record in the batch relative to FirstOffset.
				// Additionally, we must set LastOffsetDelta to the value of the last offset
				// in the batch. Since the OffsetDelta of the first record is 0, we know that the
				// final record of any batch will have an offset of (# of records in batch) - 1.
				// (See https://cwiki.apache.org/confluence/display/KAFKA/A+Guide+To+The+Kafka+Protocol#AGuideToTheKafkaProtocol-Messagesets
				//  under the RecordBatch section for details.)
				rb := set.recordsToSend.RecordBatch
				if len(rb.Records) > 0 {
					rb.LastOffsetDelta = int32(len(rb.Records) - 1)
					for i, record := range rb.Records {
						record.OffsetDelta = int64(i)
					}
				}

				req.AddBatch(topic, partition, rb)
				continue
			}
			if ps.parent.conf.Producer.Compression == CompressionNone {
				req.AddSet(topic, partition, set.recordsToSend.MsgSet)
			} else {
				// When compression is enabled, the entire set for each partition is compressed
				// and sent as the payload of a single fake "message" with the appropriate codec
				// set and no key. When the server sees a message with a compression codec, it
				// decompresses the payload and treats the result as its message set.

				if ps.parent.conf.Version.IsAtLeast(V0_10_0_0) {
					// If our version is 0.10 or later, assign relative offsets
					// to the inner messages. This lets the broker avoid
					// recompressing the message set.
					// (See https://cwiki.apache.org/confluence/display/KAFKA/KIP-31+-+Move+to+relative+offsets+in+compressed+message+sets
					// for details on relative offsets.)
					for i, msg := range set.recordsToSend.MsgSet.Messages {
						msg.Offset = int64(i)
					}
				}
				payload, err := encode(set.recordsToSend.MsgSet, ps.parent.conf.MetricRegistry)
				if err != nil {
					Logger.Println(err) // if this happens, it's basically our fault.
					panic(err)
				}
				compMsg := &Message{
					Codec:            ps.parent.conf.Producer.Compression,
					CompressionLevel: ps.parent.conf.Producer.CompressionLevel,
					Key:              nil,
					Value:            payload,
					Set:              set.recordsToSend.MsgSet, // Provide the underlying message set for accurate metrics
				}
				if ps.parent.conf.Version.IsAtLeast(V0_10_0_0) {
					compMsg.Version = 1
					compMsg.Timestamp = set.recordsToSend.MsgSet.Messages[0].Msg.Timestamp
				}
				req.AddMessage(topic, partition, compMsg)
			}
		}
	}

	return req
}

func (ps *produceSet) eachPartition(cb func(topic string, partition int32, msgs []*ProducerMessage)) {
	for topic, partitionSet := range ps.msgs {
		for partition, set := range partitionSet {
			cb(topic, partition, set.msgs)
		}
	}
}

func (ps *produceSet) dropPartition(topic string, partition int32) []*ProducerMessage {
	if ps.msgs[topic] == nil {
		return nil
	}
	set := ps.msgs[topic][partition]
	if set == nil {
		return nil
	}
	ps.bufferBytes -= set.bufferBytes
	ps.bufferCount -= len(set.msgs)
	delete(ps.msgs[topic], partition)
	return set.msgs
}

func (ps *produceSet) wouldOverflow(msg *ProducerMessage) bool {
	version := 1
	if ps.parent.conf.Version.IsAtLeast(V0_11_0_0) {
		version = 2
	}

	switch {
	// Would we overflow our maximum possible size-on-the-wire? 10KiB is arbitrary overhead for safety.
	case ps.bufferBytes+msg.byteSize(version) >= int(MaxRequestSize-(10*1024)):
		return true
	// Would we overflow the size-limit of a compressed message-batch for this partition?
	case ps.parent.conf.Producer.Compression != CompressionNone &&
		ps.msgs[msg.Topic] != nil && ps.msgs[msg.Topic][msg.Partition] != nil &&
		ps.msgs[msg.Topic][msg.Partition].bufferBytes+msg.byteSize(version) >= ps.parent.conf.Producer.MaxMessageBytes:
		return true
	// Would we overflow simply in number of messages?
	case ps.parent.conf.Producer.Flush.MaxMessages > 0 && ps.bufferCount >= ps.parent.conf.Producer.Flush.MaxMessages:
		return true
	default:
		return false
	}
}

func (ps *produceSet) readyToFlush() bool {
	switch {
	// If we don't have any messages, nothing else matters
	case ps.empty():
		return false
	// If all three config values are 0, we always flush as-fast-as-possible
	case ps.parent.conf.Producer.Flush.Frequency == 0 && ps.parent.conf.Producer.Flush.Bytes == 0 && ps.parent.conf.Producer.Flush.Messages == 0:
		return true
	// If we've passed the message trigger-point
	case ps.parent.conf.Producer.Flush.Messages > 0 && ps.bufferCount >= ps.parent.conf.Producer.Flush.Messages:
		return true
	// If we've passed the byte trigger-point
	case ps.parent.conf.Producer.Flush.Bytes > 0 && ps.bufferBytes >= ps.parent.conf.Producer.Flush.Bytes:
		return true
	default:
		return false
	}
}

func (ps *produceSet) empty() bool {
	return ps.bufferCount == 0
}
