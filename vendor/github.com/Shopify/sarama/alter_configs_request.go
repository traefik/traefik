package sarama

type AlterConfigsRequest struct {
	Resources    []*AlterConfigsResource
	ValidateOnly bool
}

type AlterConfigsResource struct {
	Type          ConfigResourceType
	Name          string
	ConfigEntries map[string]*string
}

func (acr *AlterConfigsRequest) encode(pe packetEncoder) error {
	if err := pe.putArrayLength(len(acr.Resources)); err != nil {
		return err
	}

	for _, r := range acr.Resources {
		if err := r.encode(pe); err != nil {
			return err
		}
	}

	pe.putBool(acr.ValidateOnly)
	return nil
}

func (acr *AlterConfigsRequest) decode(pd packetDecoder, version int16) error {
	resourceCount, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	acr.Resources = make([]*AlterConfigsResource, resourceCount)
	for i := range acr.Resources {
		r := &AlterConfigsResource{}
		err = r.decode(pd, version)
		if err != nil {
			return err
		}
		acr.Resources[i] = r
	}

	validateOnly, err := pd.getBool()
	if err != nil {
		return err
	}

	acr.ValidateOnly = validateOnly

	return nil
}

func (ac *AlterConfigsResource) encode(pe packetEncoder) error {
	pe.putInt8(int8(ac.Type))

	if err := pe.putString(ac.Name); err != nil {
		return err
	}

	if err := pe.putArrayLength(len(ac.ConfigEntries)); err != nil {
		return err
	}
	for configKey, configValue := range ac.ConfigEntries {
		if err := pe.putString(configKey); err != nil {
			return err
		}
		if err := pe.putNullableString(configValue); err != nil {
			return err
		}
	}

	return nil
}

func (ac *AlterConfigsResource) decode(pd packetDecoder, version int16) error {
	t, err := pd.getInt8()
	if err != nil {
		return err
	}
	ac.Type = ConfigResourceType(t)

	name, err := pd.getString()
	if err != nil {
		return err
	}
	ac.Name = name

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	if n > 0 {
		ac.ConfigEntries = make(map[string]*string, n)
		for i := 0; i < n; i++ {
			configKey, err := pd.getString()
			if err != nil {
				return err
			}
			if ac.ConfigEntries[configKey], err = pd.getNullableString(); err != nil {
				return err
			}
		}
	}
	return err
}

func (acr *AlterConfigsRequest) key() int16 {
	return 33
}

func (acr *AlterConfigsRequest) version() int16 {
	return 0
}

func (acr *AlterConfigsRequest) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
