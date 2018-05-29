package sarama

type TxnOffsetCommitRequest struct {
	TransactionalID string
	GroupID         string
	ProducerID      int64
	ProducerEpoch   int16
	Topics          map[string][]*PartitionOffsetMetadata
}

func (t *TxnOffsetCommitRequest) encode(pe packetEncoder) error {
	if err := pe.putString(t.TransactionalID); err != nil {
		return err
	}
	if err := pe.putString(t.GroupID); err != nil {
		return err
	}
	pe.putInt64(t.ProducerID)
	pe.putInt16(t.ProducerEpoch)

	if err := pe.putArrayLength(len(t.Topics)); err != nil {
		return err
	}
	for topic, partitions := range t.Topics {
		if err := pe.putString(topic); err != nil {
			return err
		}
		if err := pe.putArrayLength(len(partitions)); err != nil {
			return err
		}
		for _, partition := range partitions {
			if err := partition.encode(pe); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *TxnOffsetCommitRequest) decode(pd packetDecoder, version int16) (err error) {
	if t.TransactionalID, err = pd.getString(); err != nil {
		return err
	}
	if t.GroupID, err = pd.getString(); err != nil {
		return err
	}
	if t.ProducerID, err = pd.getInt64(); err != nil {
		return err
	}
	if t.ProducerEpoch, err = pd.getInt16(); err != nil {
		return err
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	t.Topics = make(map[string][]*PartitionOffsetMetadata)
	for i := 0; i < n; i++ {
		topic, err := pd.getString()
		if err != nil {
			return err
		}

		m, err := pd.getArrayLength()
		if err != nil {
			return err
		}

		t.Topics[topic] = make([]*PartitionOffsetMetadata, m)

		for j := 0; j < m; j++ {
			partitionOffsetMetadata := new(PartitionOffsetMetadata)
			if err := partitionOffsetMetadata.decode(pd, version); err != nil {
				return err
			}
			t.Topics[topic][j] = partitionOffsetMetadata
		}
	}

	return nil
}

func (a *TxnOffsetCommitRequest) key() int16 {
	return 28
}

func (a *TxnOffsetCommitRequest) version() int16 {
	return 0
}

func (a *TxnOffsetCommitRequest) requiredVersion() KafkaVersion {
	return V0_11_0_0
}

type PartitionOffsetMetadata struct {
	Partition int32
	Offset    int64
	Metadata  *string
}

func (p *PartitionOffsetMetadata) encode(pe packetEncoder) error {
	pe.putInt32(p.Partition)
	pe.putInt64(p.Offset)
	if err := pe.putNullableString(p.Metadata); err != nil {
		return err
	}

	return nil
}

func (p *PartitionOffsetMetadata) decode(pd packetDecoder, version int16) (err error) {
	if p.Partition, err = pd.getInt32(); err != nil {
		return err
	}
	if p.Offset, err = pd.getInt64(); err != nil {
		return err
	}
	if p.Metadata, err = pd.getNullableString(); err != nil {
		return err
	}

	return nil
}
