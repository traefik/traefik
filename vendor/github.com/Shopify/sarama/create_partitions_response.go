package sarama

import "time"

type CreatePartitionsResponse struct {
	ThrottleTime         time.Duration
	TopicPartitionErrors map[string]*TopicPartitionError
}

func (c *CreatePartitionsResponse) encode(pe packetEncoder) error {
	pe.putInt32(int32(c.ThrottleTime / time.Millisecond))
	if err := pe.putArrayLength(len(c.TopicPartitionErrors)); err != nil {
		return err
	}

	for topic, partitionError := range c.TopicPartitionErrors {
		if err := pe.putString(topic); err != nil {
			return err
		}
		if err := partitionError.encode(pe); err != nil {
			return err
		}
	}

	return nil
}

func (c *CreatePartitionsResponse) decode(pd packetDecoder, version int16) (err error) {
	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	c.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	c.TopicPartitionErrors = make(map[string]*TopicPartitionError, n)
	for i := 0; i < n; i++ {
		topic, err := pd.getString()
		if err != nil {
			return err
		}
		c.TopicPartitionErrors[topic] = new(TopicPartitionError)
		if err := c.TopicPartitionErrors[topic].decode(pd, version); err != nil {
			return err
		}
	}

	return nil
}

func (r *CreatePartitionsResponse) key() int16 {
	return 37
}

func (r *CreatePartitionsResponse) version() int16 {
	return 0
}

func (r *CreatePartitionsResponse) requiredVersion() KafkaVersion {
	return V1_0_0_0
}

type TopicPartitionError struct {
	Err    KError
	ErrMsg *string
}

func (t *TopicPartitionError) encode(pe packetEncoder) error {
	pe.putInt16(int16(t.Err))

	if err := pe.putNullableString(t.ErrMsg); err != nil {
		return err
	}

	return nil
}

func (t *TopicPartitionError) decode(pd packetDecoder, version int16) (err error) {
	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}
	t.Err = KError(kerr)

	if t.ErrMsg, err = pd.getNullableString(); err != nil {
		return err
	}

	return nil
}
