package sarama

import (
	"sort"
	"time"
)

// request message format is:
// [topic] timeout(int32)
// where topic is:
//  name(string) [partition]
// where partition is:
//  id(int32) offset(int64)

type DeleteRecordsRequest struct {
	Topics  map[string]*DeleteRecordsRequestTopic
	Timeout time.Duration
}

func (d *DeleteRecordsRequest) encode(pe packetEncoder) error {
	if err := pe.putArrayLength(len(d.Topics)); err != nil {
		return err
	}
	keys := make([]string, 0, len(d.Topics))
	for topic := range d.Topics {
		keys = append(keys, topic)
	}
	sort.Strings(keys)
	for _, topic := range keys {
		if err := pe.putString(topic); err != nil {
			return err
		}
		if err := d.Topics[topic].encode(pe); err != nil {
			return err
		}
	}
	pe.putInt32(int32(d.Timeout / time.Millisecond))

	return nil
}

func (d *DeleteRecordsRequest) decode(pd packetDecoder, version int16) error {
	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	if n > 0 {
		d.Topics = make(map[string]*DeleteRecordsRequestTopic, n)
		for i := 0; i < n; i++ {
			topic, err := pd.getString()
			if err != nil {
				return err
			}
			details := new(DeleteRecordsRequestTopic)
			if err = details.decode(pd, version); err != nil {
				return err
			}
			d.Topics[topic] = details
		}
	}

	timeout, err := pd.getInt32()
	if err != nil {
		return err
	}
	d.Timeout = time.Duration(timeout) * time.Millisecond

	return nil
}

func (d *DeleteRecordsRequest) key() int16 {
	return 21
}

func (d *DeleteRecordsRequest) version() int16 {
	return 0
}

func (d *DeleteRecordsRequest) requiredVersion() KafkaVersion {
	return V0_11_0_0
}

type DeleteRecordsRequestTopic struct {
	PartitionOffsets map[int32]int64 // partition => offset
}

func (t *DeleteRecordsRequestTopic) encode(pe packetEncoder) error {
	if err := pe.putArrayLength(len(t.PartitionOffsets)); err != nil {
		return err
	}
	keys := make([]int32, 0, len(t.PartitionOffsets))
	for partition := range t.PartitionOffsets {
		keys = append(keys, partition)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, partition := range keys {
		pe.putInt32(partition)
		pe.putInt64(t.PartitionOffsets[partition])
	}
	return nil
}

func (t *DeleteRecordsRequestTopic) decode(pd packetDecoder, version int16) error {
	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	if n > 0 {
		t.PartitionOffsets = make(map[int32]int64, n)
		for i := 0; i < n; i++ {
			partition, err := pd.getInt32()
			if err != nil {
				return err
			}
			offset, err := pd.getInt64()
			if err != nil {
				return err
			}
			t.PartitionOffsets[partition] = offset
		}
	}

	return nil
}
