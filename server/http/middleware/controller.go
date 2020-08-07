package middleware

import (
	"net/http"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/server"
)

// Controller sets the controller in the request's middleware context.
func Controller(c *controller.Controller) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := controller.CtxStore(req.Context(), c)
			req = req.WithContext(ctx)
			next.ServeHTTP(rw, req)
		})
	}
}
