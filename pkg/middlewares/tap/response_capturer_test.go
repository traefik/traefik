package tap

import (
	"bufio"
	"bytes"
	"errors"
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
	capturer := newResponseCapturer(recorder, &destination{recordBody: true, maxRecordBodySize: -1}, false)

	capturer.Header().Set("X-Backend", "yes")
	capturer.WriteHeader(http.StatusAccepted)

	_, err := capturer.Write([]byte("pong"))
	require.NoError(t, err)

	// A streamed response reaches the client as it goes.
	assert.Equal(t, http.StatusAccepted, recorder.Code)
	assert.Equal(t, "pong", recorder.Body.String())
	assert.Equal(t, "yes", recorder.Header().Get("X-Backend"))
	assert.False(t, capturer.buffering)

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
	capturer := newResponseCapturer(recorder, &destination{recordBody: true, maxRecordBodySize: -1}, true)

	capturer.Header().Set("X-Backend", "yes")
	capturer.WriteHeader(http.StatusAccepted)

	_, err := capturer.Write([]byte("pong"))
	require.NoError(t, err)

	// Nothing reaches the client before the response is served.
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Empty(t, recorder.Body.String())
	assert.Empty(t, recorder.Header().Get("X-Backend"))
	assert.True(t, capturer.buffering)

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
	assert.False(t, capturer.buffering)
}

func TestResponseCapturer_withheldRecordsHeadersSetUpstream(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set("X-Upstream", "yes")

	capturer := newResponseCapturer(recorder, &destination{recordBody: true, maxRecordBodySize: -1}, true)
	capturer.WriteHeader(http.StatusOK)

	rec := capturer.record(time.Second)
	assert.Equal(t, "yes", rec.Headers.Get("X-Upstream"))

	require.NoError(t, capturer.serve())
	assert.Equal(t, "yes", recorder.Header().Get("X-Upstream"))
}

func TestResponseCapturer_withheldTruncatesRecordOnly(t *testing.T) {
	recorder := httptest.NewRecorder()
	capturer := newResponseCapturer(recorder, &destination{recordBody: true, maxRecordBodySize: 2}, true)

	_, err := capturer.Write([]byte("pong"))
	require.NoError(t, err)

	rec := capturer.record(time.Second)
	assert.Equal(t, "po", string(rec.Body))
	assert.True(t, rec.BodyTruncated)

	require.NoError(t, capturer.serve())

	// The client still gets the whole response.
	assert.Equal(t, "pong", recorder.Body.String())
}

func TestResponseCapturer_streamedTruncatesRecord(t *testing.T) {
	testCases := []struct {
		desc              string
		maxRecordBodySize int64
		expectedBody      string
		expectedTruncated bool
	}{
		{
			desc:              "no limit",
			maxRecordBodySize: -1,
			expectedBody:      "pong!",
		},
		{
			desc:              "body at the limit",
			maxRecordBodySize: 5,
			expectedBody:      "pong!",
		},
		{
			desc:              "body over the limit",
			maxRecordBodySize: 3,
			expectedBody:      "pon",
			expectedTruncated: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			capturer := newResponseCapturer(recorder, &destination{recordBody: true, maxRecordBodySize: test.maxRecordBodySize}, false)

			for _, chunk := range []string{"po", "ng", "!"} {
				_, err := capturer.Write([]byte(chunk))
				require.NoError(t, err)
			}

			rec := capturer.record(time.Second)
			assert.Equal(t, test.expectedBody, string(rec.Body))
			assert.Equal(t, test.expectedTruncated, rec.BodyTruncated)

			// The client still gets the whole response.
			assert.Equal(t, "pong!", recorder.Body.String())
		})
	}
}

func TestResponseCapturer_bodyDisabled(t *testing.T) {
	recorder := httptest.NewRecorder()
	capturer := newResponseCapturer(recorder, &destination{recordBody: false, maxRecordBodySize: -1}, false)

	_, err := capturer.Write([]byte("pong"))
	require.NoError(t, err)

	rec := capturer.record(time.Second)
	assert.Empty(t, rec.Body)
	assert.False(t, rec.BodyTruncated)
	assert.Equal(t, "pong", recorder.Body.String())
}

func TestResponseCapturer_hijackServesWithheldResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	capturer := newResponseCapturer(recorder, &destination{recordBody: true, maxRecordBodySize: -1}, true)

	_, err := capturer.Write([]byte("pong"))
	require.NoError(t, err)

	// httptest.ResponseRecorder is not a http.Hijacker, but the withheld response is served first.
	_, _, err = capturer.Hijack()
	require.Error(t, err)
	assert.Empty(t, recorder.Body.String())

	// The response is served as soon as the underlying writer can be hijacked.
	capturer = newResponseCapturer(&hijackableRecorder{ResponseRecorder: recorder}, &destination{recordBody: true, maxRecordBodySize: -1}, true)

	_, err = capturer.Write([]byte("pong"))
	require.NoError(t, err)

	_, _, err = capturer.Hijack()
	require.Error(t, err)

	assert.Equal(t, "pong", recorder.Body.String())
	assert.False(t, capturer.buffering)

	// Once hijacked, the response is streamed.
	_, err = capturer.Write([]byte("ping"))
	require.NoError(t, err)
	assert.Equal(t, "pongping", recorder.Body.String())
}

// TestResponseCapturer_interimResponses asserts that an informational status is not taken for the
// status of the response, and is only forwarded when streaming.
func TestResponseCapturer_interimResponses(t *testing.T) {
	testCases := []struct {
		desc            string
		shouldBuffer    bool
		expectedInterim []int
		expectedFinal   []int
	}{
		{
			desc:            "streamed response",
			expectedInterim: []int{http.StatusContinue},
			expectedFinal:   []int{http.StatusContinue, http.StatusCreated},
		},
		{
			desc:          "buffered response",
			shouldBuffer:  true,
			expectedFinal: []int{http.StatusCreated},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			writer := &statusSequence{}
			capturer := newResponseCapturer(writer, &destination{recordBody: true, maxRecordBodySize: -1}, test.shouldBuffer)

			capturer.WriteHeader(http.StatusContinue)
			assert.Equal(t, test.expectedInterim, writer.codes)

			capturer.WriteHeader(http.StatusCreated)

			_, err := capturer.Write([]byte("pong"))
			require.NoError(t, err)

			rec := capturer.record(time.Second)
			assert.Equal(t, http.StatusCreated, rec.Status)
			assert.Equal(t, "pong", string(rec.Body))

			require.NoError(t, capturer.serve())

			assert.Equal(t, test.expectedFinal, writer.codes)
			assert.Equal(t, "pong", writer.body.String())
		})
	}
}

// TestResponseCapturer_bufferedInterimHeaders asserts that the headers set for a dropped informational
// response are served with the final one, as RFC 8297 expects, unless removed.
func TestResponseCapturer_bufferedInterimHeaders(t *testing.T) {
	writer := &statusSequence{}
	capturer := newResponseCapturer(writer, &destination{recordBody: true, maxRecordBodySize: -1}, true)

	capturer.Header().Add("Link", "</app.css>; rel=preload")
	capturer.Header().Set("X-Hint", "yes")
	capturer.WriteHeader(http.StatusEarlyHints)

	capturer.Header().Del("X-Hint")
	capturer.WriteHeader(http.StatusCreated)

	assert.Empty(t, writer.Header())

	require.NoError(t, capturer.serve())

	assert.Equal(t, []int{http.StatusCreated}, writer.codes)
	assert.Equal(t, []http.Header{{"Link": {"</app.css>; rel=preload"}}}, writer.sent)
}

func TestResponseCapturer_bufferedServesRemovedUpstreamHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set("X-Upstream", "yes")

	capturer := newResponseCapturer(recorder, &destination{recordBody: true, maxRecordBodySize: -1}, true)
	capturer.Header().Del("X-Upstream")
	capturer.WriteHeader(http.StatusOK)

	require.NoError(t, capturer.serve())

	assert.Empty(t, recorder.Header().Get("X-Upstream"))
}

// TestResponseCapturer_streamedTrailers asserts that a streamed response writes to the client headers
// directly, where the trailers are set once the body is written.
func TestResponseCapturer_streamedTrailers(t *testing.T) {
	recorder := httptest.NewRecorder()
	capturer := newResponseCapturer(recorder, &destination{recordBody: true, maxRecordBodySize: -1}, false)

	_, err := capturer.Write([]byte("pong"))
	require.NoError(t, err)

	capturer.Header().Set(http.TrailerPrefix+"Grpc-Status", "0")

	require.NoError(t, capturer.serve())

	assert.Equal(t, "0", recorder.Result().Trailer.Get("Grpc-Status"))
}

// hijackableRecorder is a http.Hijacker always failing to hijack, which is enough to test what comes before.
type hijackableRecorder struct {
	*httptest.ResponseRecorder
}

func (h *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("cannot hijack")
}

// statusSequence accepts informational statuses then a final one, as a server does.
// httptest.ResponseRecorder latches the first status, informational or not.
type statusSequence struct {
	header http.Header
	codes  []int
	// sent holds the headers sent along with each status.
	sent []http.Header
	body bytes.Buffer
}

func (s *statusSequence) Header() http.Header {
	if s.header == nil {
		s.header = http.Header{}
	}

	return s.header
}

func (s *statusSequence) WriteHeader(code int) {
	s.codes = append(s.codes, code)
	s.sent = append(s.sent, s.Header().Clone())
}

func (s *statusSequence) Write(p []byte) (int, error) {
	return s.body.Write(p)
}
