package authentication

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

func (a *API) handleRefreshToken(rw http.ResponseWriter, req *http.Request) {
	token, err := a.getBearerToken(rw, req)
	if err != nil {
		a.marshalErrors(rw, 401, err)
		return
	}
	ctx := req.Context()
	claims, err := a.Tokener.InspectToken(ctx, token)
	if err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}

	// Check if the claims are valid.
	if err := claims.Valid(); err != nil {
		a.marshalErrors(rw, 401, err)
		return
	}

	var refreshClaims auth.RefreshClaims
	switch ct := claims.(type) {
	case auth.RefreshClaims:
		refreshClaims = ct
	case auth.AccessClaims:
		err := httputil.ErrInvalidAuthorizationHeader()
		err.Detail = "Cannot refresh token using 'Access' token. Provide refresh token."
		a.marshalErrors(rw, 0, err)
		return
	default:
		err := httputil.ErrInternalError()
		err.Detail = "Provided unknown token claims."
		a.marshalErrors(rw, 0, err)
		return
	}
	model := mapping.NewModel(a.model)

	if err = model.SetPrimaryKeyStringValue(refreshClaims.GetAccountID()); err != nil {
		log.Debugf("Setting primary key string value failed: %v - in Refresh Token", err)
		err := httputil.ErrInternalError()
		a.marshalErrors(rw, 0, err)
		return
	}

	if err = a.serverOptions.DB.QueryCtx(req.Context(), a.model, model).Refresh(); err != nil {
		if errors.Is(err, query.ErrNoResult) {
			err = auth.ErrAccountNotFound
		}
		a.marshalErrors(rw, 0, err)
		return
	}

	tokens, err := a.Tokener.Token(model.(auth.Account))
	if err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}

	output := &LoginOutput{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    int64(tokens.ExpiresIn),
	}

	buffer := &bytes.Buffer{}
	if err = json.NewEncoder(buffer).Encode(output); err != nil {
		a.marshalErrors(rw, 500, httputil.ErrInternalError())
		return
	}
	rw.WriteHeader(http.StatusCreated)
	if _, err = buffer.WriteTo(rw); err != nil {
		log.Errorf("Writing to response writer failed: %v", err)
	}
}
