package sarama

type DeleteAclsRequest struct {
	Filters []*AclFilter
}

func (d *DeleteAclsRequest) encode(pe packetEncoder) error {
	if err := pe.putArrayLength(len(d.Filters)); err != nil {
		return err
	}

	for _, filter := range d.Filters {
		if err := filter.encode(pe); err != nil {
			return err
		}
	}

	return nil
}

func (d *DeleteAclsRequest) decode(pd packetDecoder, version int16) (err error) {
	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	d.Filters = make([]*AclFilter, n)
	for i := 0; i < n; i++ {
		d.Filters[i] = new(AclFilter)
		if err := d.Filters[i].decode(pd, version); err != nil {
			return err
		}
	}

	return nil
}

func (d *DeleteAclsRequest) key() int16 {
	return 31
}

func (d *DeleteAclsRequest) version() int16 {
	return 0
}

func (d *DeleteAclsRequest) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
