package middleware

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
)

// StoreIDFromParams stores id parameter from the httprouter params under the key: 'idKey'.
func StoreIDFromParams(idKey string) Middleware {
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
