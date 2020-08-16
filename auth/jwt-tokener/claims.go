package tokener

import (
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

// Compile time check if AccessClaims implements auth.AccessClaims.
var _ auth.AccessClaims = &AccessClaims{}

// AccessClaims is the jwt claims implementation that keeps the accountID stored in given token.
type AccessClaims struct {
	Account auth.Account `json:"account,omitempty"`
	Claims
}

// GetAccount implements auth.AccessClaims interface.
func (c *AccessClaims) GetAccount() auth.Account {
	return c.Account
}

// Compile time check if RefreshClaims implements auth.RefreshClaims.
var _ auth.RefreshClaims = &RefreshClaims{}

// RefreshClaims are the claims used for the refresh token.
type RefreshClaims struct {
	AccountID string `json:"account_id"`
	Claims
}

// GetAccountID implements auth.RefreshClaims interface.
func (r *RefreshClaims) GetAccountID() string {
	return r.AccountID
}
