package sarama

import "fmt"

const (
	unknownRecords = iota
	legacyRecords
	defaultRecords

	magicOffset = 16
	magicLength = 1
)

// Records implements a union type containing either a RecordBatch or a legacy MessageSet.
type Records struct {
	recordsType int
	msgSet      *MessageSet
	recordBatch *RecordBatch
}

func newLegacyRecords(msgSet *MessageSet) Records {
	return Records{recordsType: legacyRecords, msgSet: msgSet}
}

func newDefaultRecords(batch *RecordBatch) Records {
	return Records{recordsType: defaultRecords, recordBatch: batch}
}

// setTypeFromFields sets type of Records depending on which of msgSet or recordBatch is not nil.
// The first return value indicates whether both fields are nil (and the type is not set).
// If both fields are not nil, it returns an error.
func (r *Records) setTypeFromFields() (bool, error) {
	if r.msgSet == nil && r.recordBatch == nil {
		return true, nil
	}
	if r.msgSet != nil && r.recordBatch != nil {
		return false, fmt.Errorf("both msgSet and recordBatch are set, but record type is unknown")
	}
	r.recordsType = defaultRecords
	if r.msgSet != nil {
		r.recordsType = legacyRecords
	}
	return false, nil
}

func (r *Records) encode(pe packetEncoder) error {
	if r.recordsType == unknownRecords {
		if empty, err := r.setTypeFromFields(); err != nil || empty {
			return err
		}
	}

	switch r.recordsType {
	case legacyRecords:
		if r.msgSet == nil {
			return nil
		}
		return r.msgSet.encode(pe)
	case defaultRecords:
		if r.recordBatch == nil {
			return nil
		}
		return r.recordBatch.encode(pe)
	}
	return fmt.Errorf("unknown records type: %v", r.recordsType)
}

func (r *Records) setTypeFromMagic(pd packetDecoder) error {
	dec, err := pd.peek(magicOffset, magicLength)
	if err != nil {
		return err
	}

	magic, err := dec.getInt8()
	if err != nil {
		return err
	}

	r.recordsType = defaultRecords
	if magic < 2 {
		r.recordsType = legacyRecords
	}
	return nil
}

func (r *Records) decode(pd packetDecoder) error {
	if r.recordsType == unknownRecords {
		if err := r.setTypeFromMagic(pd); err != nil {
			return nil
		}
	}

	switch r.recordsType {
	case legacyRecords:
		r.msgSet = &MessageSet{}
		return r.msgSet.decode(pd)
	case defaultRecords:
		r.recordBatch = &RecordBatch{}
		return r.recordBatch.decode(pd)
	}
	return fmt.Errorf("unknown records type: %v", r.recordsType)
}

func (r *Records) numRecords() (int, error) {
	if r.recordsType == unknownRecords {
		if empty, err := r.setTypeFromFields(); err != nil || empty {
			return 0, err
		}
	}

	switch r.recordsType {
	case legacyRecords:
		if r.msgSet == nil {
			return 0, nil
		}
		return len(r.msgSet.Messages), nil
	case defaultRecords:
		if r.recordBatch == nil {
			return 0, nil
		}
		return len(r.recordBatch.Records), nil
	}
	return 0, fmt.Errorf("unknown records type: %v", r.recordsType)
}

func (r *Records) isPartial() (bool, error) {
	if r.recordsType == unknownRecords {
		if empty, err := r.setTypeFromFields(); err != nil || empty {
			return false, err
		}
	}

	switch r.recordsType {
	case unknownRecords:
		return false, nil
	case legacyRecords:
		if r.msgSet == nil {
			return false, nil
		}
		return r.msgSet.PartialTrailingMessage, nil
	case defaultRecords:
		if r.recordBatch == nil {
			return false, nil
		}
		return r.recordBatch.PartialTrailingRecord, nil
	}
	return false, fmt.Errorf("unknown records type: %v", r.recordsType)
}

func (r *Records) isControl() (bool, error) {
	if r.recordsType == unknownRecords {
		if empty, err := r.setTypeFromFields(); err != nil || empty {
			return false, err
		}
	}

	switch r.recordsType {
	case legacyRecords:
		return false, nil
	case defaultRecords:
		if r.recordBatch == nil {
			return false, nil
		}
		return r.recordBatch.Control, nil
	}
	return false, fmt.Errorf("unknown records type: %v", r.recordsType)
}
