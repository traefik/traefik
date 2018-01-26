package streams

import (
	"github.com/containous/traefik/middlewares/audittap/types"
	"io"
)

//-------------------------------------------------------------------------------------------------

// AuditSink interface
type AuditSink interface {
	io.Closer
	Audit(encoded types.Encoded) error
}

type noopAuditSink struct {
	types.Encoded
}

var _ AuditSink = &noopAuditSink{} // prove type conformance

func (fs *noopAuditSink) Audit(encoded types.Encoded) error {
	fs.Encoded = encoded
	return nil
}

func (fs *noopAuditSink) Close() error {
	return nil
}
