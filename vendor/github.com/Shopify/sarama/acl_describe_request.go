package sarama

type DescribeAclsRequest struct {
	AclFilter
}

func (d *DescribeAclsRequest) encode(pe packetEncoder) error {
	return d.AclFilter.encode(pe)
}

func (d *DescribeAclsRequest) decode(pd packetDecoder, version int16) (err error) {
	return d.AclFilter.decode(pd, version)
}

func (d *DescribeAclsRequest) key() int16 {
	return 29
}

func (d *DescribeAclsRequest) version() int16 {
	return 0
}

func (d *DescribeAclsRequest) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
