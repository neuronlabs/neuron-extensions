package authentication

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

func (a *API) handleRefreshToken(rw http.ResponseWriter, req *http.Request) {
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

	// Check if the claims are valid.
	if err := claims.Valid(); err != nil {
		a.marshalErrors(rw, 401, err)
		return
	}

	if _, ok := claims.(auth.AccessClaims); ok {
		err := httputil.ErrInvalidAuthorizationHeader()
		err.Detail = "Cannot refresh token using 'Access' token. Provide refresh token."
		a.marshalErrors(rw, 0, err)
		return
	}
	model := mapping.NewModel(a.model)

	if err = model.SetPrimaryKeyStringValue(claims.Subject()); err != nil {
		log.Debugf("Setting primary key string value failed: %v - in Refresh Token", err)
		err := httputil.ErrInternalError()
		a.marshalErrors(rw, 0, err)
		return
	}

	if err = a.DB.QueryCtx(req.Context(), a.model, model).Refresh(); err != nil {
		if errors.Is(err, query.ErrNoResult) {
			err = auth.ErrAccountNotFound
		}
		a.marshalErrors(rw, 0, err)
		return
	}

	tokenOptions := []auth.TokenOption{auth.TokenExpirationTime(a.Options.TokenExpiration)}
	refreshExpires := time.Unix(claims.ExpiresIn(), 0)
	// Check if the refresh token would still be valid when the
	if refreshExpires.After(time.Now().Add(a.Options.TokenExpiration)) {
		tokenOptions = append(tokenOptions, auth.TokenRefreshToken(token))
	} else {
		tokenOptions = append(tokenOptions, auth.TokenRefreshExpirationTime(a.Options.RefreshTokenExpiration))
	}

	tokenOutput, err := a.Controller.Tokener.Token(ctx, model.(auth.Account), tokenOptions...)
	if err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}

	output := &LoginOutput{
		AccessToken:  tokenOutput.AccessToken,
		RefreshToken: tokenOutput.RefreshToken,
		TokenType:    tokenOutput.TokenType,
		ExpiresIn:    int64(tokenOutput.ExpiresIn),
	}

	buffer := &bytes.Buffer{}
	if err = json.NewEncoder(buffer).Encode(output); err != nil {
		a.marshalErrors(rw, 500, httputil.ErrInternalError())
		return
	}
	a.setContentType(rw)
	rw.WriteHeader(http.StatusCreated)
	if _, err = buffer.WriteTo(rw); err != nil {
		log.Errorf("Writing to response writer failed: %v", err)
	}
}
