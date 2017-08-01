package audittap

import (
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

func TestAuditResponseWriter_with_body(t *testing.T) {
	TheClock = T0

	recorder := httptest.NewRecorder()
	w := NewAuditResponseWriter(recorder, MaximumEntityLength)
	w.WriteHeader(200)
	w.Write([]byte("hello"))
	w.Write([]byte("world"))

	ri := w.GetResponseInfo()
	assert.Equal(t, 200, ri.Status)
}

type fixedClock time.Time

func (c fixedClock) Now() time.Time {
	return time.Time(c)
}

var T0 = fixedClock(time.Unix(1000000000, 0))
