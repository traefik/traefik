package sarama

import "time"

type CreateTopicsResponse struct {
	Version      int16
	ThrottleTime time.Duration
	TopicErrors  map[string]*TopicError
}

func (c *CreateTopicsResponse) encode(pe packetEncoder) error {
	if c.Version >= 2 {
		pe.putInt32(int32(c.ThrottleTime / time.Millisecond))
	}

	if err := pe.putArrayLength(len(c.TopicErrors)); err != nil {
		return err
	}
	for topic, topicError := range c.TopicErrors {
		if err := pe.putString(topic); err != nil {
			return err
		}
		if err := topicError.encode(pe, c.Version); err != nil {
			return err
		}
	}

	return nil
}

func (c *CreateTopicsResponse) decode(pd packetDecoder, version int16) (err error) {
	c.Version = version

	if version >= 2 {
		throttleTime, err := pd.getInt32()
		if err != nil {
			return err
		}
		c.ThrottleTime = time.Duration(throttleTime) * time.Millisecond
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	c.TopicErrors = make(map[string]*TopicError, n)
	for i := 0; i < n; i++ {
		topic, err := pd.getString()
		if err != nil {
			return err
		}
		c.TopicErrors[topic] = new(TopicError)
		if err := c.TopicErrors[topic].decode(pd, version); err != nil {
			return err
		}
	}

	return nil
}

func (c *CreateTopicsResponse) key() int16 {
	return 19
}

func (c *CreateTopicsResponse) version() int16 {
	return c.Version
}

func (c *CreateTopicsResponse) requiredVersion() KafkaVersion {
	switch c.Version {
	case 2:
		return V1_0_0_0
	case 1:
		return V0_11_0_0
	default:
		return V0_10_1_0
	}
}

type TopicError struct {
	Err    KError
	ErrMsg *string
}

func (t *TopicError) encode(pe packetEncoder, version int16) error {
	pe.putInt16(int16(t.Err))

	if version >= 1 {
		if err := pe.putNullableString(t.ErrMsg); err != nil {
			return err
		}
	}

	return nil
}

func (t *TopicError) decode(pd packetDecoder, version int16) (err error) {
	kErr, err := pd.getInt16()
	if err != nil {
		return err
	}
	t.Err = KError(kErr)

	if version >= 1 {
		if t.ErrMsg, err = pd.getNullableString(); err != nil {
			return err
		}
	}

	return nil
}
