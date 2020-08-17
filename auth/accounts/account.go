package accounts

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/database"
)

//go:generate neurogns models methods --format=goimports .
//go:generate neurogns collections --format=goimports  .

// Compile time check if Account implements auth.Account interface.
var (
	_ auth.Account     = &Account{}
	_ auth.SaltGetter  = &Account{}
	_ auth.SaltSetter  = &Account{}
	_ auth.SaltFielder = &Account{}
)

// Account is the basic model used for authentication and authorization.
type Account struct {
	ID uuid.UUID
	// Timestamps for the account.
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	// Username is the unique account username.
	Username string `db:";unique"`
	// PasswordHash is the hash obtained by hashing the password.
	// Both of these fields has a json tag so that the token wouldn't keep password hash and password salt.
	PasswordHash []byte `codec:"-" json:"-"`
	PasswordSalt []byte `codec:"-" json:"-"`
}

// BeforeInsert is a hook before insertion of the account.
func (a *Account) BeforeInsert(context.Context, database.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// GetSalt implements auth.SaltGetter interface.
func (a *Account) GetSalt() []byte {
	return a.PasswordSalt
}

// SetSalt implements auth.SaltSetter interface.
func (a *Account) SetSalt(salt []byte) {
	a.PasswordSalt = salt
}

// SaltField implements auth.SaltFielder interface.
func (a *Account) SaltField() string {
	return "PasswordSalt"
}

// GetUsername implements auth.Account.
func (a *Account) GetUsername() string {
	return a.Username
}

// SetUsername implements auth.Account.
func (a *Account) SetUsername(username string) {
	a.Username = username
}

// GetPasswordHash implements auth.Account.
func (a *Account) GetPasswordHash() []byte {
	return a.PasswordHash
}

// SetPasswordHash implements auth.Account.
func (a *Account) SetPasswordHash(hash []byte) {
	a.PasswordHash = hash
}

// UsernameField implements auth.Account.
func (a *Account) UsernameField() string {
	return "Username"
}

// PasswordHashField implements auth.Account.
func (a *Account) PasswordHashField() string {
	return "PasswordHash"
}
