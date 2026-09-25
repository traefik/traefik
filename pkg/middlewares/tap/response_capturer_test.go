package tap

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseCapturer_streamed(t *testing.T) {
	recorder := httptest.NewRecorder()
	capturer := newResponseCapturer(recorder, &destination{body: true, maxBodySize: -1}, false)

	capturer.Header().Set("X-Backend", "yes")
	capturer.WriteHeader(http.StatusAccepted)

	_, err := capturer.Write([]byte("pong"))
	require.NoError(t, err)

	// A streamed response reaches the client as it goes.
	assert.Equal(t, http.StatusAccepted, recorder.Code)
	assert.Equal(t, "pong", recorder.Body.String())
	assert.Equal(t, "yes", recorder.Header().Get("X-Backend"))
	assert.True(t, capturer.served())

	capturer.Flush()
	assert.True(t, recorder.Flushed)

	require.NoError(t, capturer.serve())
	assert.Equal(t, "pong", recorder.Body.String())

	rec := capturer.record(time.Second)
	assert.Equal(t, http.StatusAccepted, rec.Status)
	assert.Equal(t, "pong", string(rec.Body))
	assert.Equal(t, time.Second, rec.Duration)
}

func TestResponseCapturer_withheld(t *testing.T) {
	recorder := httptest.NewRecorder()
	capturer := newResponseCapturer(recorder, &destination{body: true, maxBodySize: -1}, true)

	capturer.Header().Set("X-Backend", "yes")
	capturer.WriteHeader(http.StatusAccepted)

	_, err := capturer.Write([]byte("pong"))
	require.NoError(t, err)

	// Nothing reaches the client before the response is served.
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Empty(t, recorder.Body.String())
	assert.Empty(t, recorder.Header().Get("X-Backend"))
	assert.False(t, capturer.served())

	// A withheld response cannot be flushed.
	capturer.Flush()
	assert.False(t, recorder.Flushed)

	rec := capturer.record(time.Second)
	assert.Equal(t, http.StatusAccepted, rec.Status)
	assert.Equal(t, "pong", string(rec.Body))

	require.NoError(t, capturer.serve())

	assert.Equal(t, http.StatusAccepted, recorder.Code)
	assert.Equal(t, "pong", recorder.Body.String())
	assert.Equal(t, "yes", recorder.Header().Get("X-Backend"))
	assert.True(t, capturer.served())
}

func TestResponseCapturer_withheldRecordsHeadersSetUpstream(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set("X-Upstream", "yes")

	capturer := newResponseCapturer(recorder, &destination{body: true, maxBodySize: -1}, true)
	capturer.WriteHeader(http.StatusOK)

	rec := capturer.record(time.Second)
	assert.Equal(t, "yes", rec.Headers.Get("X-Upstream"))

	require.NoError(t, capturer.serve())
	assert.Equal(t, "yes", recorder.Header().Get("X-Upstream"))
}

func TestResponseCapturer_withheldTruncatesRecordOnly(t *testing.T) {
	recorder := httptest.NewRecorder()
	capturer := newResponseCapturer(recorder, &destination{body: true, maxBodySize: 2}, true)

	_, err := capturer.Write([]byte("pong"))
	require.NoError(t, err)

	rec := capturer.record(time.Second)
	assert.Equal(t, "po", string(rec.Body))
	assert.True(t, rec.BodyTruncated)

	require.NoError(t, capturer.serve())

	// The client still gets the whole response.
	assert.Equal(t, "pong", recorder.Body.String())
}

func TestResponseCapturer_bodyDisabled(t *testing.T) {
	recorder := httptest.NewRecorder()
	capturer := newResponseCapturer(recorder, &destination{body: false, maxBodySize: -1}, false)

	_, err := capturer.Write([]byte("pong"))
	require.NoError(t, err)

	rec := capturer.record(time.Second)
	assert.Empty(t, rec.Body)
	assert.False(t, rec.BodyTruncated)
	assert.Equal(t, "pong", recorder.Body.String())
}

func TestResponseCapturer_hijackServesWithheldResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	capturer := newResponseCapturer(recorder, &destination{body: true, maxBodySize: -1}, true)

	_, err := capturer.Write([]byte("pong"))
	require.NoError(t, err)

	// httptest.ResponseRecorder is not a http.Hijacker, but the withheld response is served first.
	_, _, err = capturer.Hijack()
	require.Error(t, err)
	assert.Empty(t, recorder.Body.String())

	// The response is served as soon as the underlying writer can be hijacked.
	capturer = newResponseCapturer(&hijackableRecorder{ResponseRecorder: recorder}, &destination{body: true, maxBodySize: -1}, true)

	_, err = capturer.Write([]byte("pong"))
	require.NoError(t, err)

	_, _, err = capturer.Hijack()
	require.Error(t, err)

	assert.Equal(t, "pong", recorder.Body.String())
	assert.True(t, capturer.served())

	// Once hijacked, the response is streamed.
	_, err = capturer.Write([]byte("ping"))
	require.NoError(t, err)
	assert.Equal(t, "pongping", recorder.Body.String())
}

// hijackableRecorder is a http.Hijacker that always fails to hijack,
// which is enough to assert what the capturer does before hijacking.
type hijackableRecorder struct {
	*httptest.ResponseRecorder
}

func (h *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("cannot hijack")
}

// statusSequence models what an HTTP server accepts: any number of informational
// responses, then one final status. httptest.ResponseRecorder cannot be used here, as it
// latches the first status it is given, informational or not.
type statusSequence struct {
	header http.Header
	codes  []int
	body   bytes.Buffer
}

func (s *statusSequence) Header() http.Header {
	if s.header == nil {
		s.header = http.Header{}
	}

	return s.header
}

func (s *statusSequence) WriteHeader(code int) {
	s.codes = append(s.codes, code)
}

func (s *statusSequence) Write(p []byte) (int, error) {
	return s.body.Write(p)
}

// TestResponseCapturer_interimResponses asserts that an informational response, as sent by a
// backend answering an Expect: 100-continue request, is forwarded to the client and is not
// mistaken for the status of the response.
func TestResponseCapturer_interimResponses(t *testing.T) {
	for _, withhold := range []bool{false, true} {
		t.Run(fmt.Sprintf("withhold=%v", withhold), func(t *testing.T) {
			writer := &statusSequence{}
			capturer := newResponseCapturer(writer, &destination{body: true, maxBodySize: -1}, withhold)

			capturer.WriteHeader(http.StatusContinue)

			// The interim response reaches the client even when the response is withheld:
			// the client is waiting for it before sending its body.
			require.Equal(t, []int{http.StatusContinue}, writer.codes)

			capturer.WriteHeader(http.StatusCreated)

			_, err := capturer.Write([]byte("pong"))
			require.NoError(t, err)

			rec := capturer.record(time.Second)
			assert.Equal(t, http.StatusCreated, rec.Status)
			assert.Equal(t, "pong", string(rec.Body))

			require.NoError(t, capturer.serve())

			assert.Equal(t, []int{http.StatusContinue, http.StatusCreated}, writer.codes)
			assert.Equal(t, "pong", writer.body.String())
		})
	}
}
