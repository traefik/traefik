package sarama

import "time"

type AlterConfigsResponse struct {
	ThrottleTime time.Duration
	Resources    []*AlterConfigsResourceResponse
}

type AlterConfigsResourceResponse struct {
	ErrorCode int16
	ErrorMsg  string
	Type      ConfigResourceType
	Name      string
}

func (ct *AlterConfigsResponse) encode(pe packetEncoder) error {
	pe.putInt32(int32(ct.ThrottleTime / time.Millisecond))

	if err := pe.putArrayLength(len(ct.Resources)); err != nil {
		return err
	}

	for i := range ct.Resources {
		pe.putInt16(ct.Resources[i].ErrorCode)
		err := pe.putString(ct.Resources[i].ErrorMsg)
		if err != nil {
			return nil
		}
		pe.putInt8(int8(ct.Resources[i].Type))
		err = pe.putString(ct.Resources[i].Name)
		if err != nil {
			return nil
		}
	}

	return nil
}

func (acr *AlterConfigsResponse) decode(pd packetDecoder, version int16) error {
	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	acr.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	responseCount, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	acr.Resources = make([]*AlterConfigsResourceResponse, responseCount)

	for i := range acr.Resources {
		acr.Resources[i] = new(AlterConfigsResourceResponse)

		errCode, err := pd.getInt16()
		if err != nil {
			return err
		}
		acr.Resources[i].ErrorCode = errCode

		e, err := pd.getString()
		if err != nil {
			return err
		}
		acr.Resources[i].ErrorMsg = e

		t, err := pd.getInt8()
		if err != nil {
			return err
		}
		acr.Resources[i].Type = ConfigResourceType(t)

		name, err := pd.getString()
		if err != nil {
			return err
		}
		acr.Resources[i].Name = name
	}

	return nil
}

func (r *AlterConfigsResponse) key() int16 {
	return 32
}

func (r *AlterConfigsResponse) version() int16 {
	return 0
}

func (r *AlterConfigsResponse) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
