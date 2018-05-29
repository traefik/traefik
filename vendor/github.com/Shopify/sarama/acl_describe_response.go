package sarama

import "time"

type DescribeAclsResponse struct {
	ThrottleTime time.Duration
	Err          KError
	ErrMsg       *string
	ResourceAcls []*ResourceAcls
}

func (d *DescribeAclsResponse) encode(pe packetEncoder) error {
	pe.putInt32(int32(d.ThrottleTime / time.Millisecond))
	pe.putInt16(int16(d.Err))

	if err := pe.putNullableString(d.ErrMsg); err != nil {
		return err
	}

	if err := pe.putArrayLength(len(d.ResourceAcls)); err != nil {
		return err
	}

	for _, resourceAcl := range d.ResourceAcls {
		if err := resourceAcl.encode(pe); err != nil {
			return err
		}
	}

	return nil
}

func (d *DescribeAclsResponse) decode(pd packetDecoder, version int16) (err error) {
	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	d.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}
	d.Err = KError(kerr)

	errmsg, err := pd.getString()
	if err != nil {
		return err
	}
	if errmsg != "" {
		d.ErrMsg = &errmsg
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}
	d.ResourceAcls = make([]*ResourceAcls, n)

	for i := 0; i < n; i++ {
		d.ResourceAcls[i] = new(ResourceAcls)
		if err := d.ResourceAcls[i].decode(pd, version); err != nil {
			return err
		}
	}

	return nil
}

func (d *DescribeAclsResponse) key() int16 {
	return 29
}

func (d *DescribeAclsResponse) version() int16 {
	return 0
}

func (d *DescribeAclsResponse) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
