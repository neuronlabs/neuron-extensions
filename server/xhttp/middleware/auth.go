package middleware

import (
	"net/http"
	"strings"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/server"

	"github.com/neuronlabs/neuron-extensions/server/xhttp/httputil"
)

// BearerAuthenticate gets the Authorization Header from http request, and checks if given Authorization header is valid.
func BearerAuthenticate() server.Middleware {
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
			c, ok := core.CtxGetController(ctx)
			if !ok {
				log.Errorf("No controller set in the request context. Use Controller middleware to set it up for the endpoint.")
				rw.WriteHeader(500)
				cd, ok := httputil.GetCodec(ctx)
				if ok {
					if err := cd.MarshalErrors(rw, httputil.ErrInternalError()); err != nil {
						log.Errorf("Marshaling errors failed: %v", err)
					}
				}
				return
			}
			if c.Tokener == nil {
				log.Errorf("Controller's Tokener is not defined - Bearer Authenticator requires auth.Tokener")
				cd, ok := httputil.GetCodec(ctx)
				if !ok {
					rw.WriteHeader(500)
					return
				}
				rw.Header().Set("Content-Type", cd.MimeType())
				rw.WriteHeader(500)
				if err := cd.MarshalErrors(rw, httputil.ErrInternalError()); err != nil {
					log.Errorf("Marshaling errors failed: %v", err)
				}
				return
			}
			token := strings.TrimPrefix(ah, "Bearer ")
			claims, err := c.Tokener.InspectToken(ctx, token)
			if err != nil {
				log.Debug2f("Inspecting token failed: %v", err)
				cd, ok := httputil.GetCodec(ctx)
				if !ok {
					rw.WriteHeader(http.StatusUnauthorized)
					return
				}
				rw.Header().Set("Content-Type", cd.MimeType())
				rw.WriteHeader(http.StatusUnauthorized)
				if err := cd.MarshalErrors(rw, httputil.ErrInvalidAuthenticationInfo()); err != nil {
					log.Errorf("Marshal Unauthorized error failed: %v", err)
				}
				return
			}
			switch ct := claims.(type) {
			case auth.AccessClaims:
				ctx = auth.CtxWithAccount(ctx, ct.GetAccount())
			default:
				cd, ok := httputil.GetCodec(ctx)
				if !ok {
					rw.WriteHeader(http.StatusForbidden)
					return
				}
				errAuth := httputil.ErrInvalidAuthenticationInfo()
				errAuth.Detail = "Cannot authenticate using refresh token. Refresh your token using proper endpoint."
				rw.Header().Set("Content-Type", cd.MimeType())
				rw.WriteHeader(http.StatusForbidden)
				if err := cd.MarshalErrors(rw, errAuth); err != nil {
					log.Errorf("Marshal error failed: %v", err)
				}
				return
			}
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}
