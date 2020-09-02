package tokener

import (
	"context"
	"reflect"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/store"
)

// Tokener is neuron auth.Tokener implementation for the jwt.Token.
// It allows to store and inspect token encrypted using HMAC, RSA and ECDSA algorithms.
// This structure requires store.Store to keep the revoked tokens values.
// For production ready services don't use in-memory default store.
type Tokener struct {
	Parser  jwt.Parser
	Store   store.Store
	Options auth.TokenerOptions

	signingKey, validateKey interface{}
}

// New creates new Tokener with provided 'options'.
func New(options ...auth.TokenerOption) (*Tokener, error) {
	o := &auth.TokenerOptions{
		TokenExpiration:        time.Minute * 10,
		RefreshTokenExpiration: time.Hour * 24,
		TimeFunc:               time.Now,
	}
	for _, option := range options {
		option(o)
	}

	t := &Tokener{
		Store:   o.Store,
		Parser:  jwt.Parser{SkipClaimsValidation: true},
		Options: *o,
	}

	if o.Model == nil {
		return nil, errors.Wrap(auth.ErrInitialization, "no account model defined for the tokener")
	}

	// Set the signing and validate keys for given options.
	switch t.Options.SigningMethod.(type) {
	case *jwt.SigningMethodRSA:
		if t.Options.RsaPrivateKey == nil {
			return nil, errors.Wrap(auth.ErrInvalidRSAKey, "no rsa key provided for given RSA token signing method")
		}
		t.signingKey = t.Options.RsaPrivateKey
		t.validateKey = t.Options.RsaPrivateKey.PublicKey
	case *jwt.SigningMethodHMAC:
		if len(t.Options.Secret) == 0 {
			return nil, errors.Wrap(auth.ErrInvalidSecret, "no secret provided for the HMAC token signing method")
		}
		t.signingKey, t.validateKey = t.Options.Secret, t.Options.Secret
	case *jwt.SigningMethodECDSA:
		if t.Options.EcdsaPrivateKey == nil {
			return nil, errors.Wrap(auth.ErrInvalidECDSAKey, "no ecdsa key provided for given ECDSA token signing method")
		}
		t.signingKey = t.Options.EcdsaPrivateKey
		t.validateKey = t.Options.EcdsaPrivateKey.PublicKey
	default:
		return nil, errors.Wrap(auth.ErrInitialization, "provided unsupported signing method")
	}
	return t, nil
}

// InspectToken inspects given token string and returns provided claims.
func (t *Tokener) InspectToken(ctx context.Context, token string) (auth.Claims, error) {
	claims, _, err := t.inspectToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// Token creates an auth.Token from provided options.
func (t *Tokener) Token(ctx context.Context, account auth.Account, options ...auth.TokenOption) (auth.Token, error) {
	o := &auth.TokenOptions{
		ExpirationTime:        t.Options.TokenExpiration,
		RefreshExpirationTime: t.Options.RefreshTokenExpiration,
	}
	for _, option := range options {
		option(o)
	}

	if account == nil {
		return auth.Token{}, errors.Wrap(auth.ErrNoRequiredOption, "provided no account in the token creation")
	}
	if account.IsPrimaryKeyZero() {
		return auth.Token{}, errors.Wrap(auth.ErrNoRequiredOption, "provided account with zero value primary key")
	}

	// Get string value for the account's primary key.
	accountID, err := account.GetPrimaryKeyStringValue()
	if err != nil {
		return auth.Token{}, errors.Wrapf(auth.ErrInternalError, "getting account primary key string value failed: %v", err)
	}

	// Set the claims for the full token.
	expiresAt := t.Options.TimeFunc().Add(o.ExpirationTime)
	claims := &AccessClaims{
		Account: account,
		// Set the claims with current accountID and expiresAt.
		Claims: Claims{StandardClaims: jwt.StandardClaims{Subject: accountID, ExpiresAt: expiresAt.Unix()}},
	}

	token := jwt.NewWithClaims(t.Options.SigningMethod, claims)
	tokenString, err := token.SignedString(t.signingKey)
	if err != nil {
		return auth.Token{}, errors.Wrapf(auth.ErrInternalError, "writing signed string failed: %v", err)
	}

	// Check if the refresh token is provided.
	var refreshStoreToken *StoreToken
	refreshToken := o.RefreshToken
	if refreshToken == "" {
		// Create and sign refresh token.
		refreshTokenExpiration := t.Options.TimeFunc().Add(t.Options.RefreshTokenExpiration)
		refClaims := &Claims{
			StandardClaims: jwt.StandardClaims{
				// Token subject should be account ID.
				Subject:   accountID,
				ExpiresAt: refreshTokenExpiration.Unix(),
			},
		}
		refreshTokenClaims := jwt.NewWithClaims(t.Options.SigningMethod, refClaims)
		refreshToken, err = refreshTokenClaims.SignedString(t.signingKey)
		if err != nil {
			return auth.Token{}, errors.Wrapf(auth.ErrInternalError, "writing refresh token signed string failed: %v", err)
		}
		refreshStoreToken = &StoreToken{ExpiresAt: refreshTokenExpiration}
	} else {
		refreshStoreToken, err = t.getStoreToken(ctx, refreshToken)
		if err != nil {
			return auth.Token{}, err
		}
	}
	// Add the access token to the refresh mapped tokens, and store it.
	refreshStoreToken.MappedTokens = append(refreshStoreToken.MappedTokens, tokenString)
	if err = t.setStoreToken(ctx, refreshToken, refreshStoreToken); err != nil {
		return auth.Token{}, err
	}

	// Create and set the store token for the access token with the mapped refresh token.
	sToken := &StoreToken{ExpiresAt: expiresAt, MappedTokens: []string{refreshToken}}
	if err = t.setStoreToken(ctx, tokenString, sToken); err != nil {
		return auth.Token{}, err
	}

	return auth.Token{
		AccessToken:  tokenString,
		RefreshToken: refreshToken,
		ExpiresIn:    int(o.ExpirationTime / time.Second),
		TokenType:    "bearer",
	}, nil
}

// RevokeToken invalidates provided 'token'.
func (t *Tokener) RevokeToken(ctx context.Context, token string) error {
	claims, sToken, err := t.inspectToken(ctx, token)
	if err != nil {
		return err
	}
	if sToken.RevokedAt != nil {
		return errors.Wrap(auth.ErrTokenRevoked, "token was already revoked")
	}
	if err = claims.Valid(); err != nil {
		return err
	}

	now := t.Options.TimeFunc()
	sToken.RevokedAt = &now
	if err = t.setStoreToken(ctx, token, sToken); err != nil {
		return err
	}
	var alreadyRevoked map[string]struct{}
	if len(sToken.MappedTokens) > 0 {
		alreadyRevoked = map[string]struct{}{}
	}
	for _, mappedToken := range sToken.MappedTokens {
		if err = t.revokeToken(ctx, mappedToken, now, alreadyRevoked); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tokener) revokeToken(ctx context.Context, token string, now time.Time, alreadyRevoked map[string]struct{}) error {
	if _, ok := alreadyRevoked[token]; ok {
		return nil
	}
	sToken, err := t.getStoreToken(ctx, token)
	if err != nil {
		return err
	}
	if sToken.RevokedAt != nil {
		return nil
	}
	sToken.RevokedAt = &now
	if err = t.setStoreToken(ctx, token, sToken); err != nil {
		return err
	}
	alreadyRevoked[token] = struct{}{}
	for _, mappedToken := range sToken.MappedTokens {
		if err = t.revokeToken(ctx, mappedToken, now, alreadyRevoked); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tokener) inspectToken(ctx context.Context, token string) (auth.Claims, *StoreToken, error) {
	// Initialize jwt.MapClaims.
	claims := &AccessClaims{Account: t.newAccount()}
	_, err := t.Parser.ParseWithClaims(token, claims, func(tk *jwt.Token) (interface{}, error) {
		if tk.Method != t.Options.SigningMethod {
			return nil, errors.Wrap(auth.ErrToken, "provided invalid signing algorithm for the token")
		}
		return t.validateKey, nil
	})
	if err != nil {
		if !errors.Is(err, auth.ErrToken) {
			return nil, nil, errors.Wrapf(auth.ErrToken, "parsing token failed: %v", err)
		}
		return nil, nil, err
	}
	sToken, err := t.getStoreToken(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	if sToken.RevokedAt != nil {
		claims.RevokedAt = sToken.RevokedAt.Unix()
	}

	// Check if there is account with valid ID. Otherwise set it as refresh token.
	if claims.Account.IsPrimaryKeyZero() {
		return &claims.Claims, sToken, nil
	}
	return claims, sToken, nil
}

func (t *Tokener) newAccount() auth.Account {
	tp := reflect.TypeOf(t.Options.Model)
	return reflect.New(tp.Elem()).Interface().(auth.Account)
}
