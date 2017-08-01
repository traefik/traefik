package streams

import (
	atypes "github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/containous/traefik/middlewares/audittap/types"
)

//-------------------------------------------------------------------------------------------------

type stream struct {
	sink AuditSink
}

// NewAuditStream creates a new audit stream
func NewAuditStream(sink AuditSink) atypes.AuditStream {
	return &stream{sink}
}

func (s *stream) Audit(encoder types.Encodeable) error {
	enc := encoder.ToEncoded()
	if enc.Err != nil {
		return enc.Err
	}
	return s.sink.Audit(enc)
}

func (s *stream) Close() error {
	return s.sink.Close()
}
