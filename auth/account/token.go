package account

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/errors"
)

// Claims is the jwt claims implementation that keeps the accountID stored in given token.
type Claims struct {
	AccountID uuid.UUID
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Roles     []string
	jwt.StandardClaims
}

// refreshClaims are the claims used for the refresh token.
type refreshClaims struct {
	AccountID uuid.UUID
	jwt.StandardClaims
}

// InspectToken inspects given token string and returns provided claims.
func (a *Authenticator) InspectToken(token string) (interface{}, error) {
	claims := &Claims{}
	_, err := a.Parser.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Wrap(auth.ErrToken, "provided invalid signing algorithm for the token")
		}
		return a.Options.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	if err = claims.Valid(); err != nil {
		return nil, err
	}
	return claims, nil
}

// RefreshToken generates new auth.Token based on provided refresh token.
func (a *Authenticator) RefreshToken(ctx context.Context, refreshToken string) (auth.Token, error) {
	claims := &refreshClaims{}
	_, err := a.Parser.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Wrap(auth.ErrToken, "provided invalid signing algorithm for the token")
		}
		return a.Options.Secret, nil
	})
	if err != nil {
		return auth.Token{}, err
	}
	if err = claims.Valid(); err != nil {
		return auth.Token{}, err
	}
	return a.Token(ctx, auth.TokenAccountID(claims.AccountID))
}

// Token creates an auth.Token from provided options.
func (a *Authenticator) Token(ctx context.Context, options ...auth.TokenOption) (auth.Token, error) {
	o := &auth.TokenOptions{
		ExpirationTime: a.Options.TokenExpiration,
	}
	for _, option := range options {
		option(o)
	}
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: a.c.Now().Add(o.ExpirationTime).Unix(),
		},
	}

	account, ok := CtxGetAccount(ctx)
	if !ok {
		if o.AccountID == nil {
			return auth.Token{}, errors.WrapDetf(auth.ErrAuthenticationNoRequiredOption, "provided nil account id")
		}
		var (
			id  uuid.UUID
			err error
		)
		switch accountID := o.AccountID.(type) {
		case uuid.UUID:
			id = accountID
		case string:
			id, err = uuid.Parse(accountID)
			if err != nil {
				return auth.Token{}, errors.WrapDetf(auth.ErrAuthenticationNoRequiredOption, "provided invalid uuid in the account id option")
			}
		default:
			return auth.Token{}, errors.WrapDetf(auth.ErrAuthenticationNoRequiredOption, "provided invalid account id type: %T", accountID)
		}
		// Get the account with provided id.
		account, err = NRN_Accounts.QueryCtx(ctx, a.db).
			Where("ID = ?", id).
			Get()
		if err != nil {
			return auth.Token{}, errors.WrapDetf(auth.ErrAccountNotFound, "the account with provided id is not found")
		}
	}
	claimsWithAccount(claims, account)

	token := jwt.NewWithClaims(a.SigningMethod, claims)
	tokenString, err := token.SignedString(a.Options.Secret)
	if err != nil {
		return auth.Token{}, errors.WrapDetf(errors.ErrInternal, "writing signed string failed: %v", err)
	}

	refClaims := refreshClaims{
		AccountID: account.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: a.c.Now().Add(a.Options.RefreshTokenExpiration).Unix(),
		},
	}

	refreshToken := jwt.NewWithClaims(a.SigningMethod, refClaims)
	refreshTokenString, err := refreshToken.SignedString(a.Options.Secret)
	if err != nil {
		return auth.Token{}, errors.WrapDetf(errors.ErrInternal, "writing signed string failed: %v", err)
	}
	return auth.Token{AccessToken: tokenString, RefreshToken: refreshTokenString}, nil
}

func claimsWithAccount(claims *Claims, account *Account) {
	claims.AccountID = account.ID
	claims.Email = account.Email
	claims.CreatedAt = account.CreatedAt
	claims.DeletedAt = account.DeletedAt
	claims.UpdatedAt = account.UpdatedAt
}
