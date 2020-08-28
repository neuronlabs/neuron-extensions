package httputil

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Wrap stores the httprouter params in the context.Context and handles std http handler.
func Wrap(next http.Handler) httprouter.Handle {
	return func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		req = req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, params))
		next.ServeHTTP(rw, req)
	}
}
