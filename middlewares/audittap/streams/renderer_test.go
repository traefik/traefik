package streams

import (
	"testing"
	"time"

	. "github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

type StubEncoder struct {
	enc Encoded
}

func (s *StubEncoder) ToEncoded() Encoded {
	return s.enc
}

func TestAuditStream(t *testing.T) {
	encoder := StubEncoder{Encoded{Bytes: []byte("AnyOldData")}}
	sink := &noopSink{0, 0}
	as := NewAuditStream(sink)
	err := as.Audit(&encoder)
	assert.NoError(t, err)
	assert.Equal(t, 1, sink.audits)

	err = as.Close()
	assert.NoError(t, err)
	assert.Equal(t, 1, sink.closes)
}

//-------------------------------------------------------------------------------------------------

type noopSink struct {
	audits, closes int
}

func (ns *noopSink) Audit(encoded Encoded) error {
	ns.audits++
	return nil
}

func (ns *noopSink) Close() error {
	ns.closes++
	return nil
}

//-------------------------------------------------------------------------------------------------

type fixedClock time.Time

func (c fixedClock) Now() time.Time {
	return time.Time(c)
}

var T0 = fixedClock(time.Unix(1000000000, 0))
