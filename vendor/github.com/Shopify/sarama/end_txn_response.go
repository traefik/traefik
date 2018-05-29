package sarama

import (
	"time"
)

type EndTxnResponse struct {
	ThrottleTime time.Duration
	Err          KError
}

func (e *EndTxnResponse) encode(pe packetEncoder) error {
	pe.putInt32(int32(e.ThrottleTime / time.Millisecond))
	pe.putInt16(int16(e.Err))
	return nil
}

func (e *EndTxnResponse) decode(pd packetDecoder, version int16) (err error) {
	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	e.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}
	e.Err = KError(kerr)

	return nil
}

func (e *EndTxnResponse) key() int16 {
	return 25
}

func (e *EndTxnResponse) version() int16 {
	return 0
}

func (e *EndTxnResponse) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
