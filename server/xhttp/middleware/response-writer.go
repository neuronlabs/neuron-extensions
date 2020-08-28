package middleware

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"

	"github.com/neuronlabs/brotli"

	"github.com/neuronlabs/neuron-extensions/server/xhttp/httputil"
	"github.com/neuronlabs/neuron-extensions/server/xhttp/log"
)

// ResponseWriter is a middleware that wraps the response writer into httputil.ResponseWriter.
// If the client accepts one of the supported compressions, it would also set the compressed writer.
// In order to set the default compression level set it to -1.
func ResponseWriter(compressionLevel int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			accepts := httputil.ParseAcceptEncodingHeader(req.Header)
			if len(accepts) == 0 {
				next.ServeHTTP(rw, req)
				return
			}

			buf := &bytes.Buffer{}
			customRW := &httputil.ResponseWriter{
				ResponseWriter: rw,
				// If no encoders are provided the default writer would be the buffer.
				TempWriter: buf,
				Buffer:     buf,
			}
			for _, accept := range accepts {
				thisCompression := compressionLevel
				var (
					w   io.Writer
					err error
				)
				switch accept.Value {
				case "gzip":
					switch {
					case thisCompression == -1:
						thisCompression = gzip.DefaultCompression
					case thisCompression > gzip.BestCompression:
						thisCompression = gzip.BestCompression
					case thisCompression < gzip.BestSpeed:
						thisCompression = gzip.BestSpeed
					}
					w, err = gzip.NewWriterLevel(buf, thisCompression)
				case "deflate":
					switch {
					case thisCompression == -1:
						thisCompression = flate.DefaultCompression
					case thisCompression > flate.BestCompression:
						thisCompression = flate.BestCompression
					case thisCompression < flate.BestSpeed:
						thisCompression = flate.BestSpeed
					}
					w, err = flate.NewWriter(buf, thisCompression)
				case "br":
					switch {
					case thisCompression == -1:
						thisCompression = brotli.DefaultCompression
					case thisCompression > brotli.BestCompression:
						thisCompression = brotli.BestCompression
					case thisCompression < brotli.BestSpeed:
						thisCompression = brotli.BestSpeed
					}
					w = brotli.NewWriterLevel(buf, thisCompression)
				default:
					continue
				}
				if err != nil {
					buf.Reset()
					log.Errorf("creating new response writer io.Writer failed: %v", err)
					continue
				}
				if log.Level() == log.LevelDebug3 {
					log.Debug3f("Writer: '%s' with compression level: %d", accept.Value, thisCompression)
				}
				rw.Header().Set("Content-Encoding", accept.Value)
				customRW.TempWriter = w
				break
			}
			// Serve next handlers with provided compressed writer.
			next.ServeHTTP(customRW, req)

			// Close the writer if required.
			if wc, ok := customRW.TempWriter.(io.WriteCloser); ok {
				if err := wc.Close(); err != nil {
					log.Errorf("Closing response writer failed: %v", err)
					rw.WriteHeader(500)
					return
				}
			}
			// Store the status
			rw.WriteHeader(customRW.Status)
			if _, err := buf.WriteTo(rw); err != nil {
				log.Errorf("An error occurred while writing to response writer: %v", err)
			}
			return
		})
	}
}
