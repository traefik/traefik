package fastcgi

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
)

var (
	errFastCgiProtocolError = errors.New("fastcgi request ended with error")
	fastCgiRequestEOF       = errors.New("fastcgi request complete")
)

type fastcgiReader struct {
	inner io.Reader

	contentReader io.LimitedReader
	error         bytes.Buffer
	metadata      header
}

func newFastCgiReader(src io.Reader) *fastcgiReader {
	return &fastcgiReader{
		inner: src,
		contentReader: io.LimitedReader{
			R: src,
		},
	}
}

func (r *fastcgiReader) Read(p []byte) (int, error) {
	if r.contentReader.N > 0 {
		n, err := r.contentReader.Read(p)
		if err != nil && err != io.EOF {
			return 0, err
		}

		log.Debug().Msgf("%d bytes of content read", n)
		return n, nil
	}

	if err := r.discardPadding(); err != nil {
		return 0, err
	}

	if err := r.readHeader(); err != nil {
		return 0, err
	}
	if r.metadata.Type == FastCgiEndRecord {
		return 0, r.readEndRequest()
	}

	r.contentReader.N = int64(r.metadata.ContentLength)
	switch r.metadata.Type {
	case FastCgiStderrRecord:
		_, err := io.Copy(&r.error, &r.contentReader)
		return 0, err
	case FastCgiStdoutRecord:
		return r.Read(p)
	default:
		return 0, fmt.Errorf("unexpected record type %d", r.metadata.Type)
	}
}

func (r *fastcgiReader) readHeader() error {
	var hBuff [FastCgiHeaderSz]byte
	if _, err := io.ReadFull(r.inner, hBuff[:]); err != nil {
		log.Err(err).Msg("on reading header")
		return err
	}
	if err := r.metadata.decode(hBuff[:]); err != nil {
		return err
	}
	if r.metadata.Version != FastCgiVersion {
		return errors.New("wrong fastcgi protocol version")
	}

	return nil
}

func (r *fastcgiReader) readEndRequest() error {
	var (
		req      endRequestBody
		bodyBuff [8]byte
	)

	if _, err := io.ReadFull(r.inner, bodyBuff[:]); err != nil {
		return err
	}
	if err := req.decode(bodyBuff[:]); err != nil {
		return err
	}
	if req.protocolStatus != FastCgiRequestComplete {
		return fmt.Errorf("%w: %d, %s", errFastCgiProtocolError, req.protocolStatus, req.appStatus)
	}
	if err := r.discardPadding(); err != nil {
		return err
	}

	return fastCgiRequestEOF
}

func (r *fastcgiReader) discardPadding() error {
	_, err := io.CopyN(io.Discard, r.inner, int64(r.metadata.PaddingLength))
	return err
}
