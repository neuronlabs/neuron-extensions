package account

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/db"
	"github.com/neuronlabs/neuron/errors"
)

var (
	_ auth.Authenticator = &Authenticator{}
	_ auth.Tokener       = &Authenticator{}
)

// Authenticator is the structure that implements auth.Authenticator as well as auth.Tokener interfaces.
// It is used to provide full authentication process for the
type Authenticator struct {
	Options       auth.Options
	SigningMethod jwt.SigningMethod
	Parser        jwt.Parser

	c  *controller.Controller
	db db.DB
}

// Initialize implements initializer interface.
func (a *Authenticator) Initialize(c *controller.Controller) error {
	a.c = c
	a.db = db.New(c)

	if a.SigningMethod != jwt.SigningMethodHS256 {
		return errors.NewDet(auth.ClassInitialization, "authenticator: unsupported signing method")
	}
	return nil
}

// Authenticate implements auth.Authenticator interface. Does the authentication using bcrypt algorihtm.
// Returns accountID as uuid.UUID.
func (a *Authenticator) Authenticate(ctx context.Context, email, password string) (accountID interface{}, err error) {
	account, err := Accounts.QueryCtx(ctx, a.db).Where("Email =", email).Get()
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(account.HashPassword, []byte(password)); err != nil {
		return nil, errors.NewDetf(auth.ClassInvalidSecret, "password doesn't match")
	}
	return account.ID, nil
}

// New creates new validation error.
func New(options ...auth.Option) (*Authenticator, error) {
	o := &auth.Options{
		PasswordCost:           bcrypt.DefaultCost,
		TokenExpiration:        time.Minute * 10,
		RefreshTokenExpiration: time.Hour * 24,
	}
	for _, op := range options {
		op(o)
	}

	a := &Authenticator{
		Options: *o,
		// By default use hs256 method.
		SigningMethod: jwt.SigningMethodHS256,
	}
	if err := a.validate(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Authenticator) validate() error {
	if a.Options.PasswordCost < bcrypt.MinCost || a.Options.PasswordCost > bcrypt.MaxCost {
		return errors.Newf(auth.ClassInitialization, "provided cost is out of possible values range <%d, %d>}", bcrypt.MinCost, bcrypt.MaxCost)
	}
	if a.Options.Secret == "" {
		return errors.Newf(auth.ClassInitialization, "no secret provided for the authenticator")
	}
	return nil
}
