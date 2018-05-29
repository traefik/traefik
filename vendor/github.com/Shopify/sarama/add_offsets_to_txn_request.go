package sarama

type AddOffsetsToTxnRequest struct {
	TransactionalID string
	ProducerID      int64
	ProducerEpoch   int16
	GroupID         string
}

func (a *AddOffsetsToTxnRequest) encode(pe packetEncoder) error {
	if err := pe.putString(a.TransactionalID); err != nil {
		return err
	}

	pe.putInt64(a.ProducerID)

	pe.putInt16(a.ProducerEpoch)

	if err := pe.putString(a.GroupID); err != nil {
		return err
	}

	return nil
}

func (a *AddOffsetsToTxnRequest) decode(pd packetDecoder, version int16) (err error) {
	if a.TransactionalID, err = pd.getString(); err != nil {
		return err
	}
	if a.ProducerID, err = pd.getInt64(); err != nil {
		return err
	}
	if a.ProducerEpoch, err = pd.getInt16(); err != nil {
		return err
	}
	if a.GroupID, err = pd.getString(); err != nil {
		return err
	}
	return nil
}

func (a *AddOffsetsToTxnRequest) key() int16 {
	return 25
}

func (a *AddOffsetsToTxnRequest) version() int16 {
	return 0
}

func (a *AddOffsetsToTxnRequest) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
