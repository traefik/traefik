package streams

import (
	"github.com/containous/traefik/middlewares/audittap/audittypes"
)

// Renderer is a function that encodes an audit summary.
type Renderer func(audittypes.Summary) audittypes.Encoded

//-------------------------------------------------------------------------------------------------

// DirectJSONRenderer is a Renderer that directly converts the summary to JSON.
func DirectJSONRenderer(summary audittypes.Summary) audittypes.Encoded {
	return summary.ToJSON()
}

//-------------------------------------------------------------------------------------------------

type stream struct {
	renderer Renderer
	sink     AuditSink
}

// NewAuditStream creates a new audit stream
func NewAuditStream(renderer Renderer, sink AuditSink) audittypes.AuditStream {
	return &stream{renderer, sink}
}

func (s *stream) Audit(summary audittypes.Summary) error {
	enc := s.renderer(summary)
	if enc.Err != nil {
		return enc.Err
	}
	return s.sink.Audit(enc)
}

func (s *stream) Close() error {
	return s.sink.Close()
}
