package middleware

import (
	"fmt"
	"net/http"

	"github.com/neuronlabs/neuron-plugins/server/http/httputil"
	"github.com/neuronlabs/neuron/codec"
)

// WithCodec stores the codec in the context.
func WithCodec(codecName string) Middleware {
	c, ok := codec.GetCodec(codecName)
	if !ok {
		panic(fmt.Sprintf("Codec: '%s' not found", codecName))
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			req = req.WithContext(httputil.SetCodec(req.Context(), c))
			next.ServeHTTP(rw, req)
		})
	}
}
