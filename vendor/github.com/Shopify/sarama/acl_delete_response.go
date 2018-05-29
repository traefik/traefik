package sarama

import "time"

type DeleteAclsResponse struct {
	ThrottleTime    time.Duration
	FilterResponses []*FilterResponse
}

func (a *DeleteAclsResponse) encode(pe packetEncoder) error {
	pe.putInt32(int32(a.ThrottleTime / time.Millisecond))

	if err := pe.putArrayLength(len(a.FilterResponses)); err != nil {
		return err
	}

	for _, filterResponse := range a.FilterResponses {
		if err := filterResponse.encode(pe); err != nil {
			return err
		}
	}

	return nil
}

func (a *DeleteAclsResponse) decode(pd packetDecoder, version int16) (err error) {
	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	a.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}
	a.FilterResponses = make([]*FilterResponse, n)

	for i := 0; i < n; i++ {
		a.FilterResponses[i] = new(FilterResponse)
		if err := a.FilterResponses[i].decode(pd, version); err != nil {
			return err
		}
	}

	return nil
}

func (d *DeleteAclsResponse) key() int16 {
	return 31
}

func (d *DeleteAclsResponse) version() int16 {
	return 0
}

func (d *DeleteAclsResponse) requiredVersion() KafkaVersion {
	return V0_11_0_0
}

type FilterResponse struct {
	Err          KError
	ErrMsg       *string
	MatchingAcls []*MatchingAcl
}

func (f *FilterResponse) encode(pe packetEncoder) error {
	pe.putInt16(int16(f.Err))
	if err := pe.putNullableString(f.ErrMsg); err != nil {
		return err
	}

	if err := pe.putArrayLength(len(f.MatchingAcls)); err != nil {
		return err
	}
	for _, matchingAcl := range f.MatchingAcls {
		if err := matchingAcl.encode(pe); err != nil {
			return err
		}
	}

	return nil
}

func (f *FilterResponse) decode(pd packetDecoder, version int16) (err error) {
	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}
	f.Err = KError(kerr)

	if f.ErrMsg, err = pd.getNullableString(); err != nil {
		return err
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}
	f.MatchingAcls = make([]*MatchingAcl, n)
	for i := 0; i < n; i++ {
		f.MatchingAcls[i] = new(MatchingAcl)
		if err := f.MatchingAcls[i].decode(pd, version); err != nil {
			return err
		}
	}

	return nil
}

type MatchingAcl struct {
	Err    KError
	ErrMsg *string
	Resource
	Acl
}

func (m *MatchingAcl) encode(pe packetEncoder) error {
	pe.putInt16(int16(m.Err))
	if err := pe.putNullableString(m.ErrMsg); err != nil {
		return err
	}

	if err := m.Resource.encode(pe); err != nil {
		return err
	}

	if err := m.Acl.encode(pe); err != nil {
		return err
	}

	return nil
}

func (m *MatchingAcl) decode(pd packetDecoder, version int16) (err error) {
	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}
	m.Err = KError(kerr)

	if m.ErrMsg, err = pd.getNullableString(); err != nil {
		return err
	}

	if err := m.Resource.decode(pd, version); err != nil {
		return err
	}

	if err := m.Acl.decode(pd, version); err != nil {
		return err
	}

	return nil
}
