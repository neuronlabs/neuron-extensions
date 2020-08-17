package middleware

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/neuronlabs/neuron/server"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
)

// StoreIDFromParams stores id parameter from the httprouter params under the key: 'idKey'.
func StoreIDFromParams(idKey string) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			params, ok := ctx.Value(httprouter.ParamsKey).(httprouter.Params)
			if !ok {
				log.Errorf("no httprouter.Params stored in request context")
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			ctx = httputil.CtxSetID(ctx, params.ByName(idKey))
			req = req.WithContext(ctx)
			next.ServeHTTP(rw, req)
		})
	}
}

type serverOptionsKey struct{}

var serverOptionsKeyValue = &serverOptionsKey{}

// CtxGetServerParams gets the server parameters from the context.
func CtxGetServerOptions(ctx context.Context) (*server.Options, bool) {
	params, ok := ctx.Value(serverOptionsKeyValue).(*server.Options)
	return params, ok
}

// StoreServerOptions stores the server parameters in the request context.
func StoreServerOptions(options *server.Options) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), serverOptionsKeyValue, options)
			req = req.WithContext(ctx)
			next.ServeHTTP(rw, req)
		})
	}
}
