package httputil

import (
	"bytes"
	"io"
	"net/http"
)

var _ http.ResponseWriter = &ResponseWriter{}

type ResponseWriter struct {
	Status         int
	TempWriter     io.Writer
	Buffer         *bytes.Buffer
	ResponseWriter http.ResponseWriter
}

// Header implements http.ResponseWriter.
func (r *ResponseWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

// Write implements io.Writer.
func (r *ResponseWriter) Write(bytes []byte) (int, error) {
	return r.TempWriter.Write(bytes)
}

// WriteHeader implements http.ResponseWriter.
func (r *ResponseWriter) WriteHeader(statusCode int) {
	r.Status = statusCode
}
