package sarama

import (
	"sort"
	"time"
)

// response message format is:
// throttleMs(int32) [topic]
// where topic is:
//  name(string) [partition]
// where partition is:
//  id(int32) low_watermark(int64) error_code(int16)

type DeleteRecordsResponse struct {
	Version      int16
	ThrottleTime time.Duration
	Topics       map[string]*DeleteRecordsResponseTopic
}

func (d *DeleteRecordsResponse) encode(pe packetEncoder) error {
	pe.putInt32(int32(d.ThrottleTime / time.Millisecond))

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
	return nil
}

func (d *DeleteRecordsResponse) decode(pd packetDecoder, version int16) error {
	d.Version = version

	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	d.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	if n > 0 {
		d.Topics = make(map[string]*DeleteRecordsResponseTopic, n)
		for i := 0; i < n; i++ {
			topic, err := pd.getString()
			if err != nil {
				return err
			}
			details := new(DeleteRecordsResponseTopic)
			if err = details.decode(pd, version); err != nil {
				return err
			}
			d.Topics[topic] = details
		}
	}

	return nil
}

func (d *DeleteRecordsResponse) key() int16 {
	return 21
}

func (d *DeleteRecordsResponse) version() int16 {
	return 0
}

func (d *DeleteRecordsResponse) requiredVersion() KafkaVersion {
	return V0_11_0_0
}

type DeleteRecordsResponseTopic struct {
	Partitions map[int32]*DeleteRecordsResponsePartition
}

func (t *DeleteRecordsResponseTopic) encode(pe packetEncoder) error {
	if err := pe.putArrayLength(len(t.Partitions)); err != nil {
		return err
	}
	keys := make([]int32, 0, len(t.Partitions))
	for partition := range t.Partitions {
		keys = append(keys, partition)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, partition := range keys {
		pe.putInt32(partition)
		if err := t.Partitions[partition].encode(pe); err != nil {
			return err
		}
	}
	return nil
}

func (t *DeleteRecordsResponseTopic) decode(pd packetDecoder, version int16) error {
	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	if n > 0 {
		t.Partitions = make(map[int32]*DeleteRecordsResponsePartition, n)
		for i := 0; i < n; i++ {
			partition, err := pd.getInt32()
			if err != nil {
				return err
			}
			details := new(DeleteRecordsResponsePartition)
			if err = details.decode(pd, version); err != nil {
				return err
			}
			t.Partitions[partition] = details
		}
	}

	return nil
}

type DeleteRecordsResponsePartition struct {
	LowWatermark int64
	Err          KError
}

func (t *DeleteRecordsResponsePartition) encode(pe packetEncoder) error {
	pe.putInt64(t.LowWatermark)
	pe.putInt16(int16(t.Err))
	return nil
}

func (t *DeleteRecordsResponsePartition) decode(pd packetDecoder, version int16) error {
	lowWatermark, err := pd.getInt64()
	if err != nil {
		return err
	}
	t.LowWatermark = lowWatermark

	kErr, err := pd.getInt16()
	if err != nil {
		return err
	}
	t.Err = KError(kErr)

	return nil
}
