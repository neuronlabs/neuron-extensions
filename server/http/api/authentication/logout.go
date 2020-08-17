package authentication

import (
	"net/http"
	"strings"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/errors"
)

func (a *API) handleLogout(rw http.ResponseWriter, req *http.Request) {
	token, err := a.getBearerToken(rw, req)
	if err != nil {
		a.marshalErrors(rw, 401, err)
		return
	}
	claims, err := a.Tokener.InspectToken(req.Context(), token)
	if err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}

	if err := claims.Valid(); err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}

	switch claims.(type) {
	case auth.RefreshClaims:
		err := httputil.ErrInvalidAuthorizationHeader()
		err.Detail = "provided invalid token to handleLogout"
		a.marshalErrors(rw, 0, err)
		return
	case auth.AccessClaims:
	default:
		err := httputil.ErrInternalError()
		err.Detail = "provided unknown token claims"
		a.marshalErrors(rw, 0, err)
		return
	}

	// Revoke provided token.
	err = a.Tokener.RevokeToken(req.Context(), token)
	if err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}
	return
}

func (a *API) getBearerToken(rw http.ResponseWriter, req *http.Request) (string, error) {
	header := req.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return "", errors.WrapDetf(auth.ErrAuthorizationHeader, "no bearer found").WithDetail("Authorization Header doesn't contain 'Bearer' token")

	}
	token := strings.TrimPrefix(header, "Bearer ")
	return token, nil
}
