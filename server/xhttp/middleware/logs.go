package middleware

import (
	"net/http"
	"time"

	"github.com/neuronlabs/neuron-extensions/server/xhttp/httputil"
	"github.com/neuronlabs/neuron-extensions/server/xhttp/log"
)

// LogRequest logs provided request with the response and the status.
// For full logs requires ResponseWriter middleware.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		wrapped, isWrapped := rw.(*httputil.ResponseWriter)
		if !isWrapped {
			next.ServeHTTP(rw, req)
			log.Debugf("%s %s", req.Method, req.URL.Path)
			return
		}
		ts := time.Now()
		next.ServeHTTP(rw, req)
		log.Debugf("%s %s %d %s", req.Method, req.URL.Path, wrapped.Status, time.Since(ts))
	})
}
