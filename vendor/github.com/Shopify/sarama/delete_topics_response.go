package sarama

import "time"

type DeleteTopicsResponse struct {
	Version         int16
	ThrottleTime    time.Duration
	TopicErrorCodes map[string]KError
}

func (d *DeleteTopicsResponse) encode(pe packetEncoder) error {
	if d.Version >= 1 {
		pe.putInt32(int32(d.ThrottleTime / time.Millisecond))
	}

	if err := pe.putArrayLength(len(d.TopicErrorCodes)); err != nil {
		return err
	}
	for topic, errorCode := range d.TopicErrorCodes {
		if err := pe.putString(topic); err != nil {
			return err
		}
		pe.putInt16(int16(errorCode))
	}

	return nil
}

func (d *DeleteTopicsResponse) decode(pd packetDecoder, version int16) (err error) {
	if version >= 1 {
		throttleTime, err := pd.getInt32()
		if err != nil {
			return err
		}
		d.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

		d.Version = version
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	d.TopicErrorCodes = make(map[string]KError, n)

	for i := 0; i < n; i++ {
		topic, err := pd.getString()
		if err != nil {
			return err
		}
		errorCode, err := pd.getInt16()
		if err != nil {
			return err
		}

		d.TopicErrorCodes[topic] = KError(errorCode)
	}

	return nil
}

func (d *DeleteTopicsResponse) key() int16 {
	return 20
}

func (d *DeleteTopicsResponse) version() int16 {
	return d.Version
}

func (d *DeleteTopicsResponse) requiredVersion() KafkaVersion {
	switch d.Version {
	case 1:
		return V0_11_0_0
	default:
		return V0_10_1_0
	}
}
