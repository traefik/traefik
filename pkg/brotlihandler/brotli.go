// Copyright 2013 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package brotlihandler // import "github.com/containous/traefik/v2/pkg/brotlihandler"

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

const (
	brEncoding      = "br"
	gzipEncoding    = "gzip"
	deflateEncoding = "deflate"
)

type compressResponseWriter struct {
	Writer io.WriteCloser
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier
	encoding string
	level    int
	code     int
}

func (w *compressResponseWriter) WriteHeader(c int) {
	w.ResponseWriter.Header().Del("Content-Length")
	if w.code == 0 {
		w.code = c
	}
}

func (w *compressResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *compressResponseWriter) writeHeader() {
	if w.code != 0 {
		w.ResponseWriter.WriteHeader(w.code)
		w.code = 0
	}
}

func (w *compressResponseWriter) Write(b []byte) (int, error) {
	h := w.ResponseWriter.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(b))
	}
	h.Del("Content-Length")

	e := h.Get("Content-Encoding")
	// Don't compress short pieces of data since it won't improve performance. Ignore compressed data too
	if len(b) < 1400 || e == brEncoding || e == gzipEncoding || e == deflateEncoding {
		w.writeHeader()
		return w.ResponseWriter.Write(b)
	}

	h.Set("Content-Encoding", w.encoding)
	w.writeHeader()

	if w.encoding == brEncoding {
		if w.level < brotli.BestSpeed || w.level > brotli.BestCompression {
			w.level = brotli.DefaultCompression
		}

		w.Writer = brotli.NewWriterLevel(w, w.level)
	} else {
		if w.level < gzip.HuffmanOnly || w.level > gzip.BestCompression {
			w.level = gzip.DefaultCompression
		}

		w.Writer, _ = gzip.NewWriterLevel(w, w.level)
	}
	defer w.Writer.Close()

	return w.Writer.Write(b)
}

type flusher interface {
	Flush() error
}

func (w *compressResponseWriter) Flush() {
	// Flush compressed data if compressor supports it.
	f, ok := w.Writer.(flusher)
	if !ok {
		return
	}
	f.Flush()
	// Flush HTTP response.
	if w.Flusher != nil {
		w.Flusher.Flush()
	}
}

// CompressHandler gzip/brotli compresses HTTP responses for clients that support it
// via the 'Accept-Encoding' header.
//
// Compressing TLS traffic may leak the page contents to an attacker if the
// page contains user input: http://security.stackexchange.com/a/102015/12208
func CompressHandler(h http.Handler) http.Handler {
	return CompressHandlerLevel(h, 6)
}

// CompressHandlerLevel gzip/brotli compresses HTTP responses with specified compression level
// for clients that support it via the 'Accept-Encoding' header.
//
// The compression level should be valid for encodings you use
func CompressHandlerLevel(h http.Handler, level int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// detect what encoding to use
		var encoding string
		for _, curEnc := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
			curEnc = strings.TrimSpace(curEnc)
			if curEnc == brEncoding || curEnc == gzipEncoding {
				encoding = curEnc
				if curEnc == brEncoding {
					break
				}
			}
		}

		// if we weren't able to identify an encoding we're familiar with, pass on the
		// request to the handler and return
		if encoding == "" {
			h.ServeHTTP(w, r)
			return
		}

		r.Header.Del("Accept-Encoding")
		w.Header().Add("Vary", "Accept-Encoding")

		hijacker, ok := w.(http.Hijacker)
		if !ok { /* w is not Hijacker... oh well... */
			hijacker = nil
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			flusher = nil
		}

		closeNotifier, ok := w.(http.CloseNotifier)
		if !ok {
			closeNotifier = nil
		}

		w = &compressResponseWriter{
			ResponseWriter: w,
			Hijacker:       hijacker,
			Flusher:        flusher,
			CloseNotifier:  closeNotifier,
			encoding:       encoding,
			level:          level,
		}

		h.ServeHTTP(w, r)
	})
}
