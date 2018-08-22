package sarama

type MetadataRequest struct {
	Version                int16
	Topics                 []string
	AllowAutoTopicCreation bool
}

func (r *MetadataRequest) encode(pe packetEncoder) error {
	if r.Version < 0 || r.Version > 5 {
		return PacketEncodingError{"invalid or unsupported MetadataRequest version field"}
	}
	if r.Version == 0 || r.Topics != nil || len(r.Topics) > 0 {
		err := pe.putArrayLength(len(r.Topics))
		if err != nil {
			return err
		}

		for i := range r.Topics {
			err = pe.putString(r.Topics[i])
			if err != nil {
				return err
			}
		}
	} else {
		pe.putInt32(-1)
	}
	if r.Version > 3 {
		pe.putBool(r.AllowAutoTopicCreation)
	}
	return nil
}

func (r *MetadataRequest) decode(pd packetDecoder, version int16) error {
	r.Version = version
	size, err := pd.getInt32()
	if err != nil {
		return err
	}
	if size < 0 {
		return nil
	} else {
		topicCount := size
		if topicCount == 0 {
			return nil
		}

		r.Topics = make([]string, topicCount)
		for i := range r.Topics {
			topic, err := pd.getString()
			if err != nil {
				return err
			}
			r.Topics[i] = topic
		}
	}
	if r.Version > 3 {
		autoCreation, err := pd.getBool()
		if err != nil {
			return err
		}
		r.AllowAutoTopicCreation = autoCreation
	}
	return nil
}

func (r *MetadataRequest) key() int16 {
	return 3
}

func (r *MetadataRequest) version() int16 {
	return r.Version
}

func (r *MetadataRequest) requiredVersion() KafkaVersion {
	switch r.Version {
	case 1:
		return V0_10_0_0
	case 2:
		return V0_10_1_0
	case 3, 4:
		return V0_11_0_0
	case 5:
		return V1_0_0_0
	default:
		return MinVersion
	}
}
