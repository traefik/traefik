package sarama

import "time"

type CreatePartitionsRequest struct {
	TopicPartitions map[string]*TopicPartition
	Timeout         time.Duration
	ValidateOnly    bool
}

func (c *CreatePartitionsRequest) encode(pe packetEncoder) error {
	if err := pe.putArrayLength(len(c.TopicPartitions)); err != nil {
		return err
	}

	for topic, partition := range c.TopicPartitions {
		if err := pe.putString(topic); err != nil {
			return err
		}
		if err := partition.encode(pe); err != nil {
			return err
		}
	}

	pe.putInt32(int32(c.Timeout / time.Millisecond))

	pe.putBool(c.ValidateOnly)

	return nil
}

func (c *CreatePartitionsRequest) decode(pd packetDecoder, version int16) (err error) {
	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}
	c.TopicPartitions = make(map[string]*TopicPartition, n)
	for i := 0; i < n; i++ {
		topic, err := pd.getString()
		if err != nil {
			return err
		}
		c.TopicPartitions[topic] = new(TopicPartition)
		if err := c.TopicPartitions[topic].decode(pd, version); err != nil {
			return err
		}
	}

	timeout, err := pd.getInt32()
	if err != nil {
		return err
	}
	c.Timeout = time.Duration(timeout) * time.Millisecond

	if c.ValidateOnly, err = pd.getBool(); err != nil {
		return err
	}

	return nil
}

func (r *CreatePartitionsRequest) key() int16 {
	return 37
}

func (r *CreatePartitionsRequest) version() int16 {
	return 0
}

func (r *CreatePartitionsRequest) requiredVersion() KafkaVersion {
	return V1_0_0_0
}

type TopicPartition struct {
	Count      int32
	Assignment [][]int32
}

func (t *TopicPartition) encode(pe packetEncoder) error {
	pe.putInt32(t.Count)

	if len(t.Assignment) == 0 {
		pe.putInt32(-1)
		return nil
	}

	if err := pe.putArrayLength(len(t.Assignment)); err != nil {
		return err
	}

	for _, assign := range t.Assignment {
		if err := pe.putInt32Array(assign); err != nil {
			return err
		}
	}

	return nil
}

func (t *TopicPartition) decode(pd packetDecoder, version int16) (err error) {
	if t.Count, err = pd.getInt32(); err != nil {
		return err
	}

	n, err := pd.getInt32()
	if err != nil {
		return err
	}
	if n <= 0 {
		return nil
	}
	t.Assignment = make([][]int32, n)

	for i := 0; i < int(n); i++ {
		if t.Assignment[i], err = pd.getInt32Array(); err != nil {
			return err
		}
	}

	return nil
}
