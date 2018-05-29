package sarama

type CreateAclsRequest struct {
	AclCreations []*AclCreation
}

func (c *CreateAclsRequest) encode(pe packetEncoder) error {
	if err := pe.putArrayLength(len(c.AclCreations)); err != nil {
		return err
	}

	for _, aclCreation := range c.AclCreations {
		if err := aclCreation.encode(pe); err != nil {
			return err
		}
	}

	return nil
}

func (c *CreateAclsRequest) decode(pd packetDecoder, version int16) (err error) {
	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	c.AclCreations = make([]*AclCreation, n)

	for i := 0; i < n; i++ {
		c.AclCreations[i] = new(AclCreation)
		if err := c.AclCreations[i].decode(pd, version); err != nil {
			return err
		}
	}

	return nil
}

func (d *CreateAclsRequest) key() int16 {
	return 30
}

func (d *CreateAclsRequest) version() int16 {
	return 0
}

func (d *CreateAclsRequest) requiredVersion() KafkaVersion {
	return V0_11_0_0
}

type AclCreation struct {
	Resource
	Acl
}

func (a *AclCreation) encode(pe packetEncoder) error {
	if err := a.Resource.encode(pe); err != nil {
		return err
	}
	if err := a.Acl.encode(pe); err != nil {
		return err
	}

	return nil
}

func (a *AclCreation) decode(pd packetDecoder, version int16) (err error) {
	if err := a.Resource.decode(pd, version); err != nil {
		return err
	}
	if err := a.Acl.decode(pd, version); err != nil {
		return err
	}

	return nil
}
