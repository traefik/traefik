package sarama

import (
	"time"
)

type TxnOffsetCommitResponse struct {
	ThrottleTime time.Duration
	Topics       map[string][]*PartitionError
}

func (t *TxnOffsetCommitResponse) encode(pe packetEncoder) error {
	pe.putInt32(int32(t.ThrottleTime / time.Millisecond))
	if err := pe.putArrayLength(len(t.Topics)); err != nil {
		return err
	}

	for topic, e := range t.Topics {
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

func (t *TxnOffsetCommitResponse) decode(pd packetDecoder, version int16) (err error) {
	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	t.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	t.Topics = make(map[string][]*PartitionError)

	for i := 0; i < n; i++ {
		topic, err := pd.getString()
		if err != nil {
			return err
		}

		m, err := pd.getArrayLength()
		if err != nil {
			return err
		}

		t.Topics[topic] = make([]*PartitionError, m)

		for j := 0; j < m; j++ {
			t.Topics[topic][j] = new(PartitionError)
			if err := t.Topics[topic][j].decode(pd, version); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *TxnOffsetCommitResponse) key() int16 {
	return 28
}

func (a *TxnOffsetCommitResponse) version() int16 {
	return 0
}

func (a *TxnOffsetCommitResponse) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
