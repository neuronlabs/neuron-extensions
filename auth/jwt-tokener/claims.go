package tokener

import (
	"encoding/json"

	"github.com/dgrijalva/jwt-go"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/errors"
)

// Claims is the common claims base for both AccessClaims and RefreshClaims.
// It's validation returns neuron errors, and allows to check if the token was revoked.
type Claims struct {
	RevokedAt int64 `json:"revoked_at"`
	jwt.StandardClaims
}

// Subject returns the subject of the token.
func (c *Claims) Subject() string {
	return c.StandardClaims.Subject
}

// Valid implements jwt.Claims and auth.Claims.
func (c *Claims) Valid() error {
	if c.RevokedAt != 0 {
		return auth.ErrTokenRevoked
	}
	vErr, ok := c.StandardClaims.Valid().(*jwt.ValidationError)
	if !ok {
		return nil
	}
	var multiErr errors.MultiError
	if vErr.Errors&jwt.ValidationErrorExpired != 0 {
		multiErr = append(multiErr, auth.ErrTokenExpired)
	}
	if vErr.Errors&jwt.ValidationErrorIssuedAt != 0 {
		multiErr = append(multiErr, auth.ErrToken)
	}
	if vErr.Errors&jwt.ValidationErrorNotValidYet != 0 {
		multiErr = append(multiErr, auth.ErrTokenNotValidYet)
	}
	if len(multiErr) == 0 {
		return nil
	}
	return multiErr
}

// ExpiresIn implements auth.Claims.
func (c *Claims) ExpiresIn() int64 {
	return c.ExpiresAt
}

func (c *Claims) fromMapClaims(m jwt.MapClaims) {
	// Audience
	c.Audience, _ = m["aud"].(string)
	// ExpiresAt
	switch exp := m["exp"].(type) {
	case float64:
		c.ExpiresAt = int64(exp)
	case json.Number:
		c.ExpiresAt, _ = exp.Int64()
	}
	// Set ID.
	c.Id, _ = m["id"].(string)
	// IssuedAt
	switch iat := m["iat"].(type) {
	case float64:
		c.IssuedAt = int64(iat)
	case json.Number:
		c.IssuedAt, _ = iat.Int64()
	}
	// Issuer
	c.Issuer, _ = m["iss"].(string)
	switch nbf := m["nbf"].(type) {
	case float64:
		c.NotBefore = int64(nbf)
	case json.Number:
		c.NotBefore, _ = nbf.Int64()
	}
	c.StandardClaims.Subject, _ = m["sub"].(string)
}

// Compile time check if AccessClaims implements auth.AccessClaims.
var _ auth.AccessClaims = &AccessClaims{}

// AccessClaims is the jwt claims implementation that keeps the accountID stored in given token.
type AccessClaims struct {
	Account auth.Account `json:"account"`
	Claims
}

// GetAccount implements auth.AccessClaims interface.
func (c *AccessClaims) GetAccount() auth.Account {
	return c.Account
}
