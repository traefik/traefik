package sarama

type EndTxnRequest struct {
	TransactionalID   string
	ProducerID        int64
	ProducerEpoch     int16
	TransactionResult bool
}

func (a *EndTxnRequest) encode(pe packetEncoder) error {
	if err := pe.putString(a.TransactionalID); err != nil {
		return err
	}

	pe.putInt64(a.ProducerID)

	pe.putInt16(a.ProducerEpoch)

	pe.putBool(a.TransactionResult)

	return nil
}

func (a *EndTxnRequest) decode(pd packetDecoder, version int16) (err error) {
	if a.TransactionalID, err = pd.getString(); err != nil {
		return err
	}
	if a.ProducerID, err = pd.getInt64(); err != nil {
		return err
	}
	if a.ProducerEpoch, err = pd.getInt16(); err != nil {
		return err
	}
	if a.TransactionResult, err = pd.getBool(); err != nil {
		return err
	}
	return nil
}

func (a *EndTxnRequest) key() int16 {
	return 26
}

func (a *EndTxnRequest) version() int16 {
	return 0
}

func (a *EndTxnRequest) requiredVersion() KafkaVersion {
	return V0_11_0_0
}
