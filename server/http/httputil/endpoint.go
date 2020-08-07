package httputil

import (
	"context"
	"net/http"

	"github.com/neuronlabs/neuron/server"
)

var endpointCtxKey = &endpointKey{}

type endpointKey struct{}

// MidStoreEndpoint is a middleware that stores the endpoint in the request context.
func MidStoreEndpoint(endpoint *server.Endpoint) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), endpointCtxKey, endpoint)
			req = req.WithContext(ctx)
			next.ServeHTTP(rw, req)
		})
	}
}

// CtxGetEndpoint gets the endpoint stored in the given 'ctx' context.
func CtxGetEndpoint(ctx context.Context) (*server.Endpoint, bool) {
	endpoint, ok := ctx.Value(endpointCtxKey).(*server.Endpoint)
	return endpoint, ok
}
