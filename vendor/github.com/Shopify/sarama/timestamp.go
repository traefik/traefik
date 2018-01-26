package sarama

import (
	"fmt"
	"time"
)

type Timestamp struct {
	*time.Time
}

func (t Timestamp) encode(pe packetEncoder) error {
	timestamp := int64(-1)

	if !t.Before(time.Unix(0, 0)) {
		timestamp = t.UnixNano() / int64(time.Millisecond)
	} else if !t.IsZero() {
		return PacketEncodingError{fmt.Sprintf("invalid timestamp (%v)", t)}
	}

	pe.putInt64(timestamp)
	return nil
}

func (t Timestamp) decode(pd packetDecoder) error {
	millis, err := pd.getInt64()
	if err != nil {
		return err
	}

	// negative timestamps are invalid, in these cases we should return
	// a zero time
	timestamp := time.Time{}
	if millis >= 0 {
		timestamp = time.Unix(millis/1000, (millis%1000)*int64(time.Millisecond))
	}

	*t.Time = timestamp
	return nil
}
