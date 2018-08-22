package sarama

import "time"

type DeleteTopicsRequest struct {
	Version int16
	Topics  []string
	Timeout time.Duration
}

func (d *DeleteTopicsRequest) encode(pe packetEncoder) error {
	if err := pe.putStringArray(d.Topics); err != nil {
		return err
	}
	pe.putInt32(int32(d.Timeout / time.Millisecond))

	return nil
}

func (d *DeleteTopicsRequest) decode(pd packetDecoder, version int16) (err error) {
	if d.Topics, err = pd.getStringArray(); err != nil {
		return err
	}
	timeout, err := pd.getInt32()
	if err != nil {
		return err
	}
	d.Timeout = time.Duration(timeout) * time.Millisecond
	d.Version = version
	return nil
}

func (d *DeleteTopicsRequest) key() int16 {
	return 20
}

func (d *DeleteTopicsRequest) version() int16 {
	return d.Version
}

func (d *DeleteTopicsRequest) requiredVersion() KafkaVersion {
	switch d.Version {
	case 1:
		return V0_11_0_0
	default:
		return V0_10_1_0
	}
}
