package timetools

import (
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// We use RFC2822 format for timestamps everywhere ('Thu, 13 Oct 2011 18:02:00 GMT'), but
// by default Go's JSON package uses another format when decoding/encoding timestamps.
// That's why we declare our own timestamp type and implement its marshal/unmarshal methods
// so JSON package decodes/encodes it properly.
type RFC2822Time time.Time

func NewRFC2822Time(timestamp int64) RFC2822Time {
	return RFC2822Time(time.Unix(timestamp, 0).UTC())
}

func (t RFC2822Time) Unix() int64 {
	return time.Time(t).Unix()
}

func (t RFC2822Time) IsZero() bool {
	return time.Time(t).IsZero()
}

func (t RFC2822Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(time.Time(t).Format(time.RFC1123))), nil
}

func (t *RFC2822Time) UnmarshalJSON(s []byte) error {
	q, err := strconv.Unquote(string(s))
	if err != nil {
		return err
	}
	if *(*time.Time)(t), err = time.Parse(time.RFC1123, q); err != nil {
		return err
	}
	return nil
}

func (t RFC2822Time) GetBSON() (interface{}, error) {
	return time.Time(t), nil
}

func (t *RFC2822Time) SetBSON(raw bson.Raw) error {
	var result time.Time
	err := raw.Unmarshal(&result)
	if err != nil {
		return err
	}
	*t = RFC2822Time(result)
	return nil
}

func (t RFC2822Time) String() string {
	return time.Time(t).Format(time.RFC1123)
}
