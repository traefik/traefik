package sarama

import (
	"time"
)

type AddPartitionsToTxnResponse struct {
	ThrottleTime time.Duration
	Errors       map[string][]*PartitionError
}

func (a *AddPartitionsToTxnResponse) encode(pe packetEncoder) error {
	pe.putInt32(int32(a.ThrottleTime / time.Millisecond))
	if err := pe.putArrayLength(len(a.Errors)); err != nil {
		return err
	}

	for topic, e := range a.Errors {
		if err := pe.putString(topic); err != nil {
			return err
		}
		if err := pe.putArrayLength(len(e)); err != nil {
			return err
		}
		for _, partitionError := range e {
			if err := partitionError.encode(pe); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *AddPartitionsToTxnResponse) decode(pd packetDecoder, version int16) (err error) {
	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	a.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	a.Errors = make(map[string][]*PartitionError)

	for i := 0; i < n; i++ {
		topic, err := pd.getString()
		if err != nil {
			return err
		}

		m, err := pd.getArrayLength()
		if err != nil {
			return err
		}

		a.Errors[topic] = make([]*PartitionError, m)

		for j := 0; j < m; j++ {
			a.Errors[topic][j] = new(PartitionError)
			if err := a.Errors[topic][j].decode(pd, version); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *AddPartitionsToTxnResponse) key() int16 {
	return 24
}

func (a *AddPartitionsToTxnResponse) version() int16 {
	return 0
}

func (a *AddPartitionsToTxnResponse) requiredVersion() KafkaVersion {
	return V0_11_0_0
}

type PartitionError struct {
	Partition int32
	Err       KError
}

func (p *PartitionError) encode(pe packetEncoder) error {
	pe.putInt32(p.Partition)
	pe.putInt16(int16(p.Err))
	return nil
}

func (p *PartitionError) decode(pd packetDecoder, version int16) (err error) {
	if p.Partition, err = pd.getInt32(); err != nil {
		return err
	}

	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}
	p.Err = KError(kerr)

	return nil
}
