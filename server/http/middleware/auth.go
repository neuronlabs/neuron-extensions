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
			ctx := req.Context()
			if !strings.HasPrefix(ah, "Bearer") {
				rw.WriteHeader(http.StatusUnauthorized)
				cd, ok := httputil.GetCodec(ctx)
				if ok {
					if err := cd.MarshalErrors(rw, httputil.ErrInvalidAuthorizationHeader()); err != nil {
						log.Errorf("Marshal Unauthorized error failed: %v", err)
					}
				}
				return
			}

			claims, err := tokener.InspectToken(ctx, ah)
			if err != nil {
				rw.WriteHeader(http.StatusUnauthorized)
				cd, ok := httputil.GetCodec(ctx)
				if ok {
					if err := cd.MarshalErrors(rw, httputil.ErrInvalidAuthenticationInfo()); err != nil {
						log.Errorf("Marshal Unauthorized error failed: %v", err)
					}
				}
				return
			}
			switch ct := claims.(type) {
			case auth.AccessClaims:
				ctx = auth.CtxWithAccount(ctx, ct.GetAccount())
			case auth.RefreshClaims:
				rw.WriteHeader(http.StatusForbidden)
				cd, ok := httputil.GetCodec(ctx)
				if ok {
					errAuth := httputil.ErrInvalidAuthenticationInfo()
					errAuth.Detail = "Cannot authenticate using refresh token. Refresh your token using proper endpoint."
					if err := cd.MarshalErrors(rw, errAuth); err != nil {
						log.Errorf("Marshal error failed: %v", err)
					}
				}
			}
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}
