package sarama

type AddPartitionsToTxnRequest struct {
	TransactionalID string
	ProducerID      int64
	ProducerEpoch   int16
	TopicPartitions map[string][]int32
}

func (a *AddPartitionsToTxnRequest) encode(pe packetEncoder) error {
	if err := pe.putString(a.TransactionalID); err != nil {
		return err
	}
	pe.putInt64(a.ProducerID)
	pe.putInt16(a.ProducerEpoch)

	if err := pe.putArrayLength(len(a.TopicPartitions)); err != nil {
		return err
	}
	for topic, partitions := range a.TopicPartitions {
		if err := pe.putString(topic); err != nil {
			return err
		}
		if err := pe.putInt32Array(partitions); err != nil {
			return err
		}
	}

	return nil
}

func (a *AddPartitionsToTxnRequest) decode(pd packetDecoder, version int16) (err error) {
	if a.TransactionalID, err = pd.getString(); err != nil {
		return err
	}
	if a.ProducerID, err = pd.getInt64(); err != nil {
		return err
	}
	if a.ProducerEpoch, err = pd.getInt16(); err != nil {
		return err
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	a.TopicPartitions = make(map[string][]int32)
	for i := 0; i < n; i++ {
		topic, err := pd.getString()
		if err != nil {
			return err
		}

		partitions, err := pd.getInt32Array()
		if err != nil {
			return err
		}

		a.TopicPartitions[topic] = partitions
	}

	return nil
}

func (a *AddPartitionsToTxnRequest) key() int16 {
	return 24
}

func (a *AddPartitionsToTxnRequest) version() int16 {
	return 0
}

func (a *AddPartitionsToTxnRequest) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
