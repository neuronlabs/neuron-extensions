package middleware

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"

	"github.com/neuronlabs/brotli"
	"github.com/neuronlabs/jsonapi-handler/log"

	"github.com/neuronlabs/neuron-plugins/server/http/httputil"
)

var _ http.ResponseWriter = &responseWriter{}

type responseWriter struct {
	rw http.ResponseWriter
	w  io.Writer
}

// Header implements http.ResponseWriter.
func (r responseWriter) Header() http.Header {
	return r.rw.Header()
}

// Write implements io.Writer.
func (r responseWriter) Write(bytes []byte) (int, error) {
	return r.w.Write(bytes)
}

// WriteHeader implements http.ResponseWriter.
func (r responseWriter) WriteHeader(statusCode int) {
	r.rw.WriteHeader(statusCode)
}

func ResponseWriter(compressionLevel int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			accepts := httputil.ParseAcceptEncodingHeader(req.Header)
			if len(accepts) == 0 {
				next.ServeHTTP(rw, req)
				return
			}

			w := io.Writer(rw)
			var err error

			for _, accept := range accepts {
				switch accept.Value {
				case "gzip":
					switch {
					case compressionLevel > gzip.BestCompression:
						compressionLevel = gzip.BestCompression
					case compressionLevel < gzip.BestSpeed:
						compressionLevel = gzip.BestSpeed
					case compressionLevel == -1:
						compressionLevel = gzip.DefaultCompression
					}
					w, err = gzip.NewWriterLevel(rw, compressionLevel)
				case "deflate":
					switch {
					case compressionLevel > flate.BestCompression:
						compressionLevel = flate.BestCompression
					case compressionLevel < flate.BestSpeed:
						compressionLevel = flate.BestSpeed
					case compressionLevel == -1:
						compressionLevel = flate.DefaultCompression
					}
					w, err = flate.NewWriter(rw, compressionLevel)
				case "br":
					switch {
					case compressionLevel > brotli.BestCompression:
						compressionLevel = brotli.BestCompression
					case compressionLevel < brotli.BestSpeed:
						compressionLevel = brotli.BestSpeed
					case compressionLevel == -1:
						compressionLevel = brotli.DefaultCompression
					}
					w = brotli.NewWriterLevel(rw, compressionLevel)
				default:
					continue
				}
				if err != nil {

				}
				if log.Level() == log.LDEBUG3 {
					log.Debug3f("Writer: '%s' with compression level: %d", accept.Value, compressionLevel)
				}
				rw.Header().Set("Content-Encoding", accept.Value)
				next.ServeHTTP(&responseWriter{rw: rw, w: w}, req)
				if wc, ok := w.(io.WriteCloser); ok {
					if err = wc.Close(); err != nil {
						log.Errorf("Closing response writer failed: %v", err)
					}
				}
				return
			}
			next.ServeHTTP(rw, req)
		})
	}
}
