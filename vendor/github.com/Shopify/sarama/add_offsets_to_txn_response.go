package sarama

import (
	"time"
)

type AddOffsetsToTxnResponse struct {
	ThrottleTime time.Duration
	Err          KError
}

func (a *AddOffsetsToTxnResponse) encode(pe packetEncoder) error {
	pe.putInt32(int32(a.ThrottleTime / time.Millisecond))
	pe.putInt16(int16(a.Err))
	return nil
}

func (a *AddOffsetsToTxnResponse) decode(pd packetDecoder, version int16) (err error) {
	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	a.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}
	a.Err = KError(kerr)

	return nil
}

func (a *AddOffsetsToTxnResponse) key() int16 {
	return 25
}

func (a *AddOffsetsToTxnResponse) version() int16 {
	return 0
}

func (a *AddOffsetsToTxnResponse) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
