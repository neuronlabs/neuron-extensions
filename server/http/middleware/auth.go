package middleware

import (
	"net/http"
	"strings"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/server"
)

// BearerAuthenticate gets the Authorization Header from http request, and checks if given Authorization header is valid.
func BearerAuthenticate(tokener auth.Tokener) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ah := req.Header.Get("Authorization")
			if !strings.HasPrefix(ah, "Bearer") {
				rw.WriteHeader(http.StatusUnauthorized)
				cd, ok := httputil.GetCodec(req.Context())
				if ok {
					if err := cd.MarshalErrors(rw, httputil.ErrInvalidAuthorizationHeader()); err != nil {
						log.Errorf("Marshal Unauthorized error failed: %v", err)
					}
				}
				return
			}
			accountID, err := tokener.InspectToken(ah)
			if err != nil {
				rw.WriteHeader(http.StatusUnauthorized)
				cd, ok := httputil.GetCodec(req.Context())
				if ok {
					if err := cd.MarshalErrors(rw, httputil.ErrInvalidAuthenticationInfo()); err != nil {
						log.Errorf("Marshal Unauthorized error failed: %v", err)
					}
				}
				return
			}
			ctx := auth.CtxWithAccountID(req.Context(), accountID)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}
