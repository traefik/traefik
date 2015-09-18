// Copyright 2013 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handlers

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressResponseWriter struct {
	io.Writer
	http.ResponseWriter
	http.Hijacker
}

func (w *compressResponseWriter) WriteHeader(c int) {
	w.ResponseWriter.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(c)
}

func (w *compressResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *compressResponseWriter) Write(b []byte) (int, error) {
	h := w.ResponseWriter.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(b))
	}
	h.Del("Content-Length")

	return w.Writer.Write(b)
}

// CompressHandler gzip compresses HTTP responses for clients that support it
// via the 'Accept-Encoding' header.
func CompressHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	L:
		for _, enc := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
			switch strings.TrimSpace(enc) {
			case "gzip":
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Add("Vary", "Accept-Encoding")

				gw := gzip.NewWriter(w)
				defer gw.Close()

				h, hok := w.(http.Hijacker)
				if !hok { /* w is not Hijacker... oh well... */
					h = nil
				}

				w = &compressResponseWriter{
					Writer:         gw,
					ResponseWriter: w,
					Hijacker:       h,
				}

				break L
			case "deflate":
				w.Header().Set("Content-Encoding", "deflate")
				w.Header().Add("Vary", "Accept-Encoding")

				fw, _ := flate.NewWriter(w, flate.DefaultCompression)
				defer fw.Close()

				h, hok := w.(http.Hijacker)
				if !hok { /* w is not Hijacker... oh well... */
					h = nil
				}

				w = &compressResponseWriter{
					Writer:         fw,
					ResponseWriter: w,
					Hijacker:       h,
				}

				break L
			}
		}

		h.ServeHTTP(w, r)
	})
}
