// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package httpstat

import (
	"bufio"
	"net"
	"net/http"
)

// ResponseWriter is a custom implementation of the http.ResponseWriter
// interface.
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher

	// Status returns the status code of the response or 0 if the response
	// has not been written.
	Status() int
	// Written returns whether or not the ResponseWriter has been written to.
	Written() bool
	// BytesWritten returns the amount of bytes written to the response body.
	BytesWritten() int
}

type responseRecorder struct {
	http.ResponseWriter

	status       int
	bytesWritten int
}

// NewResponseRecorder returns a new instance of a responseRecorder.
func NewResponseRecorder(w http.ResponseWriter) ResponseWriter {
	return &responseRecorder{ResponseWriter: w}
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *responseRecorder) Status() int {
	return r.status
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet.
		r.WriteHeader(http.StatusOK)
	}

	bytesWritten, err := r.ResponseWriter.Write(b)
	r.bytesWritten += bytesWritten
	return bytesWritten, err
}

func (r *responseRecorder) BytesWritten() int {
	return r.bytesWritten
}

func (r *responseRecorder) Written() bool {
	return r.status != 0
}

func (r *responseRecorder) CloseNotify() <-chan bool {
	notifier, ok := r.ResponseWriter.(http.CloseNotifier)
	if !ok {
		panic("wrapped ResponseWriter does not support the CloseNotifier interface")
	}
	return notifier.CloseNotify()
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		panic("wrapped ResponseWriter does not support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func approxRequestSize(r *http.Request) (size int) {
	// References to " + 2" are for newlines.

	// Use RequestURI rather than URI.String(), to cut down on additional
	// processing.
	if r.RequestURI != "" {
		size = len(r.RequestURI)
	} else {
		size = len(r.URL.RequestURI())
	}

	size += len(r.Method) + 2
	size += len(r.Proto) + 2
	if r.Host != "" {
		// "Host: " + "\r\n".
		size += len(r.Host) + 6 + 2
	}

	for name, values := range r.Header {
		size += len(name) + 2

		for _, value := range values {
			size += len(value)
		}
	}

	if len(r.TransferEncoding) > 0 {
		// "Transfer-Encoding: " + "\r\n".
		size += 19 + 2
		for _, enc := range r.TransferEncoding {
			size += len(enc) + 1
		}

		// Subtract length for no trailing comma.
		size--
	}

	if r.Close {
		// "Connection: close\r\n".
		size += 17 + 2

	}

	// Form and MultipartForm should be included in the serialized URL.
	if r.ContentLength > 0 {
		// Newline between headers and body.
		size += 2
		size += int(r.ContentLength)
	}

	return size
}
