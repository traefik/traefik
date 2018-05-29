package sarama

type ConfigResource struct {
	Type        ConfigResourceType
	Name        string
	ConfigNames []string
}

type DescribeConfigsRequest struct {
	Resources []*ConfigResource
}

func (r *DescribeConfigsRequest) encode(pe packetEncoder) error {
	if err := pe.putArrayLength(len(r.Resources)); err != nil {
		return err
	}

	for _, c := range r.Resources {
		pe.putInt8(int8(c.Type))
		if err := pe.putString(c.Name); err != nil {
			return err
		}

		if len(c.ConfigNames) == 0 {
			pe.putInt32(-1)
			continue
		}
		if err := pe.putStringArray(c.ConfigNames); err != nil {
			return err
		}
	}

	return nil
}

func (r *DescribeConfigsRequest) decode(pd packetDecoder, version int16) (err error) {
	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	r.Resources = make([]*ConfigResource, n)

	for i := 0; i < n; i++ {
		r.Resources[i] = &ConfigResource{}
		t, err := pd.getInt8()
		if err != nil {
			return err
		}
		r.Resources[i].Type = ConfigResourceType(t)
		name, err := pd.getString()
		if err != nil {
			return err
		}
		r.Resources[i].Name = name

		confLength, err := pd.getArrayLength()

		if err != nil {
			return err
		}

		if confLength == -1 {
			continue
		}

		cfnames := make([]string, confLength)
		for i := 0; i < confLength; i++ {
			s, err := pd.getString()
			if err != nil {
				return err
			}
			cfnames[i] = s
		}
		r.Resources[i].ConfigNames = cfnames
	}

	return nil
}

func (r *DescribeConfigsRequest) key() int16 {
	return 32
}

func (r *DescribeConfigsRequest) version() int16 {
	return 0
}

func (r *DescribeConfigsRequest) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
