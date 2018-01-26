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

func (s *stream) Audit(enc types.Encoded) error {
	return s.sink.Audit(enc)
}

func (s *stream) Close() error {
	return s.sink.Close()
}
