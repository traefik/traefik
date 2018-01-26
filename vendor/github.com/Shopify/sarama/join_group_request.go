package sarama

type GroupProtocol struct {
	Name     string
	Metadata []byte
}

func (p *GroupProtocol) decode(pd packetDecoder) (err error) {
	p.Name, err = pd.getString()
	if err != nil {
		return err
	}
	p.Metadata, err = pd.getBytes()
	return err
}

func (p *GroupProtocol) encode(pe packetEncoder) (err error) {
	if err := pe.putString(p.Name); err != nil {
		return err
	}
	if err := pe.putBytes(p.Metadata); err != nil {
		return err
	}
	return nil
}

type JoinGroupRequest struct {
	GroupId               string
	SessionTimeout        int32
	MemberId              string
	ProtocolType          string
	GroupProtocols        map[string][]byte // deprecated; use OrderedGroupProtocols
	OrderedGroupProtocols []*GroupProtocol
}

func (r *JoinGroupRequest) encode(pe packetEncoder) error {
	if err := pe.putString(r.GroupId); err != nil {
		return err
	}
	pe.putInt32(r.SessionTimeout)
	if err := pe.putString(r.MemberId); err != nil {
		return err
	}
	if err := pe.putString(r.ProtocolType); err != nil {
		return err
	}

	if len(r.GroupProtocols) > 0 {
		if len(r.OrderedGroupProtocols) > 0 {
			return PacketDecodingError{"cannot specify both GroupProtocols and OrderedGroupProtocols on JoinGroupRequest"}
		}

		if err := pe.putArrayLength(len(r.GroupProtocols)); err != nil {
			return err
		}
		for name, metadata := range r.GroupProtocols {
			if err := pe.putString(name); err != nil {
				return err
			}
			if err := pe.putBytes(metadata); err != nil {
				return err
			}
		}
	} else {
		if err := pe.putArrayLength(len(r.OrderedGroupProtocols)); err != nil {
			return err
		}
		for _, protocol := range r.OrderedGroupProtocols {
			if err := protocol.encode(pe); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *JoinGroupRequest) decode(pd packetDecoder, version int16) (err error) {
	if r.GroupId, err = pd.getString(); err != nil {
		return
	}

	if r.SessionTimeout, err = pd.getInt32(); err != nil {
		return
	}

	if r.MemberId, err = pd.getString(); err != nil {
		return
	}

	if r.ProtocolType, err = pd.getString(); err != nil {
		return
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}

	r.GroupProtocols = make(map[string][]byte)
	for i := 0; i < n; i++ {
		protocol := &GroupProtocol{}
		if err := protocol.decode(pd); err != nil {
			return err
		}
		r.GroupProtocols[protocol.Name] = protocol.Metadata
		r.OrderedGroupProtocols = append(r.OrderedGroupProtocols, protocol)
	}

	return nil
}

func (r *JoinGroupRequest) key() int16 {
	return 11
}

func (r *JoinGroupRequest) version() int16 {
	return 0
}

func (r *JoinGroupRequest) requiredVersion() KafkaVersion {
	return V0_9_0_0
}

func (r *JoinGroupRequest) AddGroupProtocol(name string, metadata []byte) {
	r.OrderedGroupProtocols = append(r.OrderedGroupProtocols, &GroupProtocol{
		Name:     name,
		Metadata: metadata,
	})
}

func (r *JoinGroupRequest) AddGroupProtocolMetadata(name string, metadata *ConsumerGroupMemberMetadata) error {
	bin, err := encode(metadata, nil)
	if err != nil {
		return err
	}

	r.AddGroupProtocol(name, bin)
	return nil
}
