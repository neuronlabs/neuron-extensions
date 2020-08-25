package authentication

import (
	"net/http"
	"strings"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/errors"
)

func (a *API) handleLogout(rw http.ResponseWriter, req *http.Request) {
	token, err := a.getBearerToken(req)
	if err != nil {
		a.marshalErrors(rw, 401, err)
		return
	}
	ctx := req.Context()
	claims, err := a.Controller.Tokener.InspectToken(ctx, token)
	if err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}

	// Checks if given claims are still valid.
	if err := claims.Valid(); err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}

	if !a.Options.PermitRefreshTokenLogout {
		if _, ok := claims.(auth.AccessClaims); !ok {
			err := httputil.ErrInvalidAuthorizationHeader()
			err.Detail = "provided invalid token to log out"
			a.marshalErrors(rw, 0, err)
			return
		}
	}

	// Revoke provided token.
	err = a.Controller.Tokener.RevokeToken(ctx, token)
	if err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}
	rw.WriteHeader(http.StatusNoContent)
}

func (a *API) getBearerToken(req *http.Request) (string, error) {
	header := req.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return "", errors.WrapDetf(auth.ErrAuthorizationHeader, "no bearer found").WithDetail("Authorization Header doesn't contain 'Bearer' token")

	}
	token := strings.TrimPrefix(header, "Bearer ")
	return token, nil
}
