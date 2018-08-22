package sarama

type DeleteGroupsRequest struct {
	Groups []string
}

func (r *DeleteGroupsRequest) encode(pe packetEncoder) error {
	return pe.putStringArray(r.Groups)
}

func (r *DeleteGroupsRequest) decode(pd packetDecoder, version int16) (err error) {
	r.Groups, err = pd.getStringArray()
	return
}

func (r *DeleteGroupsRequest) key() int16 {
	return 42
}

func (r *DeleteGroupsRequest) version() int16 {
	return 0
}

func (r *DeleteGroupsRequest) requiredVersion() KafkaVersion {
	return V1_1_0_0
}

func (r *DeleteGroupsRequest) AddGroup(group string) {
	r.Groups = append(r.Groups, group)
}
