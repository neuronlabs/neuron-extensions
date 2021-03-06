package authenticator

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"golang.org/x/crypto/bcrypt"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/errors"
)

var (
	_ auth.Authenticator = &Authenticator{}
)

// Authenticator is the structure that implements auth.Authenticator as well as auth.Tokener interfaces.
// It is used to provide full authentication process for the
type Authenticator struct {
	Options *auth.AuthenticatorOptions
}

// HashAndSetPassword implements auth.PasswordHasher interface.
func (a *Authenticator) HashAndSetPassword(acc auth.Account, password *auth.Password) error {
	switch a.Options.AuthenticateMethod {
	case auth.BCrypt:
		return a.setBCryptPassword(acc, password)
	default:
		return a.setHashedPassword(acc, password)
	}
}

// ComparePassword compares the password (with optional salt) hash with the provided 'password'.
func (a *Authenticator) ComparePassword(acc auth.Account, password string) error {
	if a.Options.AuthenticateMethod == auth.BCrypt {
		err := bcrypt.CompareHashAndPassword(acc.GetPasswordHash(), []byte(password))
		if err != nil {
			return errors.Wrap(auth.ErrInvalidPassword, "passwords doesn't match")
		}
		return nil
	}
	// The password is based on the hash + salt.
	var h hash.Hash
	switch a.Options.AuthenticateMethod {
	case auth.MD5:
		h = md5.New()
	case auth.SHA256:
		h = sha256.New()
	case auth.SHA512:
		h = sha512.New()
	default:
		return errors.Wrap(auth.ErrInternalError, "unsupported authentication method")
	}
	var salt []byte
	if salter, ok := acc.(auth.SaltGetter); ok {
		salt = salter.GetSalt()
	}
	match, err := auth.CompareHashPassword(h, password, acc.GetPasswordHash(), salt)
	if err != nil {
		return err
	}
	if match {
		return nil
	}
	return errors.Wrap(auth.ErrInvalidPassword, "password doesn't match")
}

// New creates new authenticator for provided options.
// By default it uses in-memory store for the revoked tokens.
func New(options ...auth.AuthenticatorOption) *Authenticator {
	o := &auth.AuthenticatorOptions{
		SaltLength: 10,
		BCryptCost: bcrypt.DefaultCost,
	}
	for _, op := range options {
		op(o)
	}
	return &Authenticator{
		Options: o,
	}
}

func (a *Authenticator) validate() error {
	// Check if provided authentication method is valid.
	switch a.Options.AuthenticateMethod {
	case auth.BCrypt:
		if a.Options.BCryptCost < bcrypt.MinCost || a.Options.BCryptCost > bcrypt.MaxCost {
			return errors.Wrapf(auth.ErrInitialization, "provided cost is out of possible values range <%d, %d>}", bcrypt.MinCost, bcrypt.MaxCost)
		}
	case auth.MD5, auth.SHA256, auth.SHA512:
		if a.Options.SaltLength == 0 {
			return errors.Wrap(auth.ErrInitialization, "provided 0 value for the authentication salt length")
		}
	default:
		return errors.Wrap(auth.ErrInitialization, "provided unknown authentication method")
	}
	return nil
}

func (a *Authenticator) setBCryptPassword(acc auth.Account, password *auth.Password) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password.Password), a.Options.BCryptCost)
	if err != nil {
		return errors.Wrapf(auth.ErrInternalError, "generating password with bcrypt failed: %v", err)
	}
	acc.SetPasswordHash(hashed)
	return nil
}

func (a *Authenticator) setHashedPassword(acc auth.Account, password *auth.Password) error {
	var (
		salt []byte
		err  error
	)
	if a.Options.SaltLength != 0 {
		salt, err = auth.GenerateSalt(a.Options.SaltLength)
		if err != nil {
			return err
		}
		if saltSetter, ok := acc.(auth.SaltSetter); ok {
			saltSetter.SetSalt(salt)
		}
	}
	var h hash.Hash
	switch a.Options.AuthenticateMethod {
	case auth.MD5:
		h = md5.New()
	case auth.SHA256:
		h = sha256.New()
	case auth.SHA512:
		h = sha512.New()
	}
	hashed, err := password.Hash(h, salt)
	if err != nil {
		return err
	}
	acc.SetPasswordHash(hashed)
	return nil
}
