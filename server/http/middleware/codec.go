package middleware

import (
	"net/http"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/server"
)

// WithCodec stores the codec in the context.
func WithCodec(c codec.Codec) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			req = req.WithContext(httputil.SetCodec(req.Context(), c))
			next.ServeHTTP(rw, req)
		})
	}
}
