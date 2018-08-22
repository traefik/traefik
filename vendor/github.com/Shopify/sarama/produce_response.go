package sarama

import (
	"fmt"
	"time"
)

type ProduceResponseBlock struct {
	Err    KError
	Offset int64
	// only provided if Version >= 2 and the broker is configured with `LogAppendTime`
	Timestamp time.Time
}

func (b *ProduceResponseBlock) decode(pd packetDecoder, version int16) (err error) {
	tmp, err := pd.getInt16()
	if err != nil {
		return err
	}
	b.Err = KError(tmp)

	b.Offset, err = pd.getInt64()
	if err != nil {
		return err
	}

	if version >= 2 {
		if millis, err := pd.getInt64(); err != nil {
			return err
		} else if millis != -1 {
			b.Timestamp = time.Unix(millis/1000, (millis%1000)*int64(time.Millisecond))
		}
	}

	return nil
}

func (b *ProduceResponseBlock) encode(pe packetEncoder, version int16) (err error) {
	pe.putInt16(int16(b.Err))
	pe.putInt64(b.Offset)

	if version >= 2 {
		timestamp := int64(-1)
		if !b.Timestamp.Before(time.Unix(0, 0)) {
			timestamp = b.Timestamp.UnixNano() / int64(time.Millisecond)
		} else if !b.Timestamp.IsZero() {
			return PacketEncodingError{fmt.Sprintf("invalid timestamp (%v)", b.Timestamp)}
		}
		pe.putInt64(timestamp)
	}

	return nil
}

type ProduceResponse struct {
	Blocks       map[string]map[int32]*ProduceResponseBlock
	Version      int16
	ThrottleTime time.Duration // only provided if Version >= 1
}

func (r *ProduceResponse) decode(pd packetDecoder, version int16) (err error) {
	r.Version = version

	numTopics, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	r.Blocks = make(map[string]map[int32]*ProduceResponseBlock, numTopics)
	for i := 0; i < numTopics; i++ {
		name, err := pd.getString()
		if err != nil {
			return err
		}

		numBlocks, err := pd.getArrayLength()
		if err != nil {
			return err
		}

		r.Blocks[name] = make(map[int32]*ProduceResponseBlock, numBlocks)

		for j := 0; j < numBlocks; j++ {
			id, err := pd.getInt32()
			if err != nil {
				return err
			}

			block := new(ProduceResponseBlock)
			err = block.decode(pd, version)
			if err != nil {
				return err
			}
			r.Blocks[name][id] = block
		}
	}

	if r.Version >= 1 {
		millis, err := pd.getInt32()
		if err != nil {
			return err
		}

		r.ThrottleTime = time.Duration(millis) * time.Millisecond
	}

	return nil
}

func (r *ProduceResponse) encode(pe packetEncoder) error {
	err := pe.putArrayLength(len(r.Blocks))
	if err != nil {
		return err
	}
	for topic, partitions := range r.Blocks {
		err = pe.putString(topic)
		if err != nil {
			return err
		}
		err = pe.putArrayLength(len(partitions))
		if err != nil {
			return err
		}
		for id, prb := range partitions {
			pe.putInt32(id)
			err = prb.encode(pe, r.Version)
			if err != nil {
				return err
			}
		}
	}
	if r.Version >= 1 {
		pe.putInt32(int32(r.ThrottleTime / time.Millisecond))
	}
	return nil
}

func (r *ProduceResponse) key() int16 {
	return 0
}

func (r *ProduceResponse) version() int16 {
	return r.Version
}

func (r *ProduceResponse) requiredVersion() KafkaVersion {
	switch r.Version {
	case 1:
		return V0_9_0_0
	case 2:
		return V0_10_0_0
	case 3:
		return V0_11_0_0
	default:
		return MinVersion
	}
}

func (r *ProduceResponse) GetBlock(topic string, partition int32) *ProduceResponseBlock {
	if r.Blocks == nil {
		return nil
	}

	if r.Blocks[topic] == nil {
		return nil
	}

	return r.Blocks[topic][partition]
}

// Testing API

func (r *ProduceResponse) AddTopicPartition(topic string, partition int32, err KError) {
	if r.Blocks == nil {
		r.Blocks = make(map[string]map[int32]*ProduceResponseBlock)
	}
	byTopic, ok := r.Blocks[topic]
	if !ok {
		byTopic = make(map[int32]*ProduceResponseBlock)
		r.Blocks[topic] = byTopic
	}
	byTopic[partition] = &ProduceResponseBlock{Err: err}
}
