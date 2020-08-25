package middleware

import (
	"net/http"

	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/server"
)

// Controller sets the controller in the request's middleware context.
func Controller(c *core.Controller) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := core.CtxSetController(req.Context(), c)
			req = req.WithContext(ctx)
			next.ServeHTTP(rw, req)
		})
	}
}
