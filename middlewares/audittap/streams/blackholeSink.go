package streams

import (
	"github.com/containous/traefik/middlewares/audittap/types"
	"io"
	"io/ioutil"
)

type blackholeSink struct {
	w io.Writer
}

// NewBlackholeSink creates a new sink that discards all data
func NewBlackholeSink() AuditSink {
	return &blackholeSink{ioutil.Discard}
}

func (s *blackholeSink) Audit(encoded types.Encoded) error {
	_, err := s.w.Write(encoded.Bytes)
	return err
}

func (s *blackholeSink) Close() error {
	return nil // Noop
}
