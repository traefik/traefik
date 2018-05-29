package sarama

import (
	"time"
)

type CreateTopicsRequest struct {
	Version int16

	TopicDetails map[string]*TopicDetail
	Timeout      time.Duration
	ValidateOnly bool
}

func (c *CreateTopicsRequest) encode(pe packetEncoder) error {
	if err := pe.putArrayLength(len(c.TopicDetails)); err != nil {
		return err
	}
	for topic, detail := range c.TopicDetails {
		if err := pe.putString(topic); err != nil {
			return err
		}
		if err := detail.encode(pe); err != nil {
			return err
		}
	}

	pe.putInt32(int32(c.Timeout / time.Millisecond))

	if c.Version >= 1 {
		pe.putBool(c.ValidateOnly)
	}

	return nil
}

func (c *CreateTopicsRequest) decode(pd packetDecoder, version int16) (err error) {
	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	c.TopicDetails = make(map[string]*TopicDetail, n)

	for i := 0; i < n; i++ {
		topic, err := pd.getString()
		if err != nil {
			return err
		}
		c.TopicDetails[topic] = new(TopicDetail)
		if err = c.TopicDetails[topic].decode(pd, version); err != nil {
			return err
		}
	}

	timeout, err := pd.getInt32()
	if err != nil {
		return err
	}
	c.Timeout = time.Duration(timeout) * time.Millisecond

	if version >= 1 {
		c.ValidateOnly, err = pd.getBool()
		if err != nil {
			return err
		}

		c.Version = version
	}

	return nil
}

func (c *CreateTopicsRequest) key() int16 {
	return 19
}

func (c *CreateTopicsRequest) version() int16 {
	return c.Version
}

func (c *CreateTopicsRequest) requiredVersion() KafkaVersion {
	switch c.Version {
	case 2:
		return V1_0_0_0
	case 1:
		return V0_11_0_0
	default:
		return V0_10_1_0
	}
}

type TopicDetail struct {
	NumPartitions     int32
	ReplicationFactor int16
	ReplicaAssignment map[int32][]int32
	ConfigEntries     map[string]*string
}

func (t *TopicDetail) encode(pe packetEncoder) error {
	pe.putInt32(t.NumPartitions)
	pe.putInt16(t.ReplicationFactor)

	if err := pe.putArrayLength(len(t.ReplicaAssignment)); err != nil {
		return err
	}
	for partition, assignment := range t.ReplicaAssignment {
		pe.putInt32(partition)
		if err := pe.putInt32Array(assignment); err != nil {
			return err
		}
	}

	if err := pe.putArrayLength(len(t.ConfigEntries)); err != nil {
		return err
	}
	for configKey, configValue := range t.ConfigEntries {
		if err := pe.putString(configKey); err != nil {
			return err
		}
		if err := pe.putNullableString(configValue); err != nil {
			return err
		}
	}

	return nil
}

func (t *TopicDetail) decode(pd packetDecoder, version int16) (err error) {
	if t.NumPartitions, err = pd.getInt32(); err != nil {
		return err
	}
	if t.ReplicationFactor, err = pd.getInt16(); err != nil {
		return err
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	if n > 0 {
		t.ReplicaAssignment = make(map[int32][]int32, n)
		for i := 0; i < n; i++ {
			replica, err := pd.getInt32()
			if err != nil {
				return err
			}
			if t.ReplicaAssignment[replica], err = pd.getInt32Array(); err != nil {
				return err
			}
		}
	}

	n, err = pd.getArrayLength()
	if err != nil {
		return err
	}

	if n > 0 {
		t.ConfigEntries = make(map[string]*string, n)
		for i := 0; i < n; i++ {
			configKey, err := pd.getString()
			if err != nil {
				return err
			}
			if t.ConfigEntries[configKey], err = pd.getNullableString(); err != nil {
				return err
			}
		}
	}

	return nil
}
