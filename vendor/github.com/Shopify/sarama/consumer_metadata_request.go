package sarama

type ConsumerMetadataRequest struct {
	ConsumerGroup string
}

func (r *ConsumerMetadataRequest) encode(pe packetEncoder) error {
	tmp := new(FindCoordinatorRequest)
	tmp.CoordinatorKey = r.ConsumerGroup
	tmp.CoordinatorType = CoordinatorGroup
	return tmp.encode(pe)
}

func (r *ConsumerMetadataRequest) decode(pd packetDecoder, version int16) (err error) {
	tmp := new(FindCoordinatorRequest)
	if err := tmp.decode(pd, version); err != nil {
		return err
	}
	r.ConsumerGroup = tmp.CoordinatorKey
	return nil
}

func (r *ConsumerMetadataRequest) key() int16 {
	return 10
}

func (r *ConsumerMetadataRequest) version() int16 {
	return 0
}

func (r *ConsumerMetadataRequest) requiredVersion() KafkaVersion {
	return V0_8_2_0
}
