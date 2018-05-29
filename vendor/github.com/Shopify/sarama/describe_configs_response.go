package sarama

import "time"

type DescribeConfigsResponse struct {
	ThrottleTime time.Duration
	Resources    []*ResourceResponse
}

type ResourceResponse struct {
	ErrorCode int16
	ErrorMsg  string
	Type      ConfigResourceType
	Name      string
	Configs   []*ConfigEntry
}

type ConfigEntry struct {
	Name      string
	Value     string
	ReadOnly  bool
	Default   bool
	Sensitive bool
}

func (r *DescribeConfigsResponse) encode(pe packetEncoder) (err error) {
	pe.putInt32(int32(r.ThrottleTime / time.Millisecond))
	if err = pe.putArrayLength(len(r.Resources)); err != nil {
		return err
	}

	for _, c := range r.Resources {
		if err = c.encode(pe); err != nil {
			return err
		}
	}
	return nil
}

func (r *DescribeConfigsResponse) decode(pd packetDecoder, version int16) (err error) {
	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	r.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	r.Resources = make([]*ResourceResponse, n)
	for i := 0; i < n; i++ {
		rr := &ResourceResponse{}
		if err := rr.decode(pd, version); err != nil {
			return err
		}
		r.Resources[i] = rr
	}

	return nil
}

func (r *DescribeConfigsResponse) key() int16 {
	return 32
}

func (r *DescribeConfigsResponse) version() int16 {
	return 0
}

func (r *DescribeConfigsResponse) requiredVersion() KafkaVersion {
	return V0_11_0_0
}

func (r *ResourceResponse) encode(pe packetEncoder) (err error) {
	pe.putInt16(r.ErrorCode)

	if err = pe.putString(r.ErrorMsg); err != nil {
		return err
	}

	pe.putInt8(int8(r.Type))

	if err = pe.putString(r.Name); err != nil {
		return err
	}

	if err = pe.putArrayLength(len(r.Configs)); err != nil {
		return err
	}

	for _, c := range r.Configs {
		if err = c.encode(pe); err != nil {
			return err
		}
	}
	return nil
}

func (r *ResourceResponse) decode(pd packetDecoder, version int16) (err error) {
	ec, err := pd.getInt16()
	if err != nil {
		return err
	}
	r.ErrorCode = ec

	em, err := pd.getString()
	if err != nil {
		return err
	}
	r.ErrorMsg = em

	t, err := pd.getInt8()
	if err != nil {
		return err
	}
	r.Type = ConfigResourceType(t)

	name, err := pd.getString()
	if err != nil {
		return err
	}
	r.Name = name

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	r.Configs = make([]*ConfigEntry, n)
	for i := 0; i < n; i++ {
		c := &ConfigEntry{}
		if err := c.decode(pd, version); err != nil {
			return err
		}
		r.Configs[i] = c
	}
	return nil
}

func (r *ConfigEntry) encode(pe packetEncoder) (err error) {
	if err = pe.putString(r.Name); err != nil {
		return err
	}

	if err = pe.putString(r.Value); err != nil {
		return err
	}

	pe.putBool(r.ReadOnly)
	pe.putBool(r.Default)
	pe.putBool(r.Sensitive)
	return nil
}

func (r *ConfigEntry) decode(pd packetDecoder, version int16) (err error) {
	name, err := pd.getString()
	if err != nil {
		return err
	}
	r.Name = name

	value, err := pd.getString()
	if err != nil {
		return err
	}
	r.Value = value

	read, err := pd.getBool()
	if err != nil {
		return err
	}
	r.ReadOnly = read

	de, err := pd.getBool()
	if err != nil {
		return err
	}
	r.Default = de

	sensitive, err := pd.getBool()
	if err != nil {
		return err
	}
	r.Sensitive = sensitive
	return nil
}
