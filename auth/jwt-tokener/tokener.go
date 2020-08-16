package tokener

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
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

	c                       *controller.Controller
	signingKey, validateKey interface{}
}

// New creates new Tokener with provided 'options'.
func New(options ...auth.TokenerOption) (*Tokener, error) {
	o := &auth.TokenerOptions{
		TokenExpiration:        time.Minute * 10,
		RefreshTokenExpiration: time.Hour * 24,
	}
	for _, option := range options {
		option(o)
	}

	t := &Tokener{
		Parser:  jwt.Parser{SkipClaimsValidation: true},
		Options: *o,
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

// Initialize implements core.Initializer interface.
func (t *Tokener) Initialize(c *controller.Controller) error {
	t.c = c

	if t.Store == nil {
		if c.DefaultStore == nil {
			return errors.Wrap(auth.ErrInitialization, "no store found for the authenticator")
		}
		t.Store = c.DefaultStore
	}
	return nil
}

// InspectToken inspects given token string and returns provided claims.
func (t *Tokener) InspectToken(ctx context.Context, token string) (auth.Claims, error) {
	mapClaims, err := t.inspectToken(ctx, token)
	if err != nil {
		return nil, err
	}

	var claims auth.Claims
	_, isAccess := mapClaims["account"]
	if !isAccess {
		if _, ok := mapClaims["account_id"]; !ok {
			return nil, errors.Wrap(auth.ErrToken, "provided token with invalid claims")
		}
		claims = &RefreshClaims{}
	} else {
		claims = &AccessClaims{}
	}
	marshaled, err := json.Marshal(mapClaims)
	if err != nil {
		return nil, errors.Wrap(auth.ErrInternalError, "marshaling map claims failed")
	}
	if err = json.Unmarshal(marshaled, claims); err != nil {
		return nil, errors.Wrap(auth.ErrInternalError, "unmarshaling claims failed")
	}
	return claims, nil
}

// Token creates an auth.Token from provided options.
func (t *Tokener) Token(account auth.Account, options ...auth.TokenOption) (auth.Token, error) {
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

	// Set the claims for the full token.
	claims := &AccessClaims{
		Account: account,
		Claims: Claims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: t.c.Now().Add(o.ExpirationTime).Unix(),
			},
		},
	}

	token := jwt.NewWithClaims(t.Options.SigningMethod, claims)

	tokenString, err := token.SignedString(t.signingKey)
	if err != nil {
		return auth.Token{}, errors.Wrapf(auth.ErrInternalError, "writing signed string failed: %v", err)
	}

	// Get string value for the account's primary key.
	stringID, err := account.GetPrimaryKeyStringValue()
	if err != nil {
		return auth.Token{}, errors.Wrapf(auth.ErrInternalError, "getting account primary key string value failed: %v", err)
	}

	// Create and sign refresh token.
	refClaims := &RefreshClaims{
		AccountID: stringID,
		Claims: Claims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: t.c.Now().Add(t.Options.RefreshTokenExpiration).Unix(),
			},
		},
	}

	refreshToken := jwt.NewWithClaims(t.Options.SigningMethod, refClaims)
	refreshTokenString, err := refreshToken.SignedString(t.signingKey)
	if err != nil {
		return auth.Token{}, errors.Wrapf(auth.ErrInternalError, "writing signed string failed: %v", err)
	}
	return auth.Token{
		AccessToken:  tokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int(o.ExpirationTime / time.Second),
		TokenType:    "bearer",
	}, nil
}

// RevokeToken invalidates provided 'token'.
func (t *Tokener) RevokeToken(ctx context.Context, token string) error {
	claims := Claims{}
	if _, err := t.inspectToken(ctx, token); err != nil {
		return err
	}
	now := jwt.TimeFunc().Unix()
	ttl := time.Unix(now, 0).Sub(time.Unix(claims.ExpiresAt, 0))

	value := make([]byte, 8)
	binary.BigEndian.PutUint64(value, uint64(now))
	err := t.Store.SetWithTTL(ctx, &store.Record{Key: t.revokeKey(token), Value: value, ExpiresAt: t.c.Now().Add(ttl)}, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (t *Tokener) inspectToken(ctx context.Context, token string) (jwt.MapClaims, error) {
	// Initialize jwt.MapClaims.
	claims := jwt.MapClaims{}
	_, err := t.Parser.ParseWithClaims(token, claims, func(tk *jwt.Token) (interface{}, error) {
		if tk.Method != t.Options.SigningMethod {
			return nil, errors.Wrap(auth.ErrToken, "provided invalid signing algorithm for the token")
		}
		return t.validateKey, nil
	})
	if err != nil {
		if !errors.Is(err, auth.ErrToken) {
			return nil, errors.Wrapf(auth.ErrToken, "parsing token failed: %v", err)
		}
		return nil, err
	}

	// Check if the token is not set as revoked.
	record, err := t.Store.Get(ctx, t.revokeKey(token))
	if err != nil {
		// If the token was not revoked than the error would be of store.ErrValueNotFound.
		if errors.Is(err, store.ErrRecordNotFound) {
			return claims, nil
		}
		log.Errorf("Getting token info from store failed: %v", err)
		return nil, err
	}

	// Set the revoked at field.
	revokedAt := binary.BigEndian.Uint64(record.Value)
	// The store had marked this token as revoked.
	claims["revoked_at"] = revokedAt
	return claims, errors.Wrap(auth.ErrTokenRevoked, "provided token had been revoked")
}

func (t *Tokener) revokeKey(token string) string {
	return "nrn_jwt_auth_revoked-" + token
}
