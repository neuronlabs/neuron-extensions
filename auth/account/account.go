package account

import (
	"time"

	"github.com/google/uuid"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

//go:generate neuron-generator models --format=goimports --single-file .
//go:generate neuron-generator collections --format=goimports  --single-file .

// Account is the basic model used for authentication and authorization.
type Account struct {
	ID uuid.UUID
	// Timestamps for the account.
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	// Email is the unique
	Email string `db:"unique,notnull"`
	// HashPassword is the hash obtained by hashing the password.
	HashPassword []byte `codec:"-"`
	// Password is the user provided password.
	Password string `db:"-"`
	Roles    []*Role
}

// Validate does the validation of the input account. It checks the email address as well as the
func (a *Account) InsertValidate() error {
	if a.Email == "" {
		return errors.New(mapping.ClassFieldValue, "provided empty email value")
	}
	return nil
}
