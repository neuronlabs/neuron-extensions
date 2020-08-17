package auth

import (
	"context"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query/filter"
)

// CheckUsernameHandler is an interface that allows to check the username in a custom way.
type CheckUsernameHandler interface {
	HandleCheckUsername(ctx context.Context, db database.DB, account auth.Account) error
}

// AccountGetHandler is an interface that handles getting account.
type AccountGetterHandler interface {
	HandleGetAccount(ctx context.Context, db database.DB, account auth.Account) error
}

// DefaultHandler is the default model handler for the accounts.
type DefaultHandler struct {
	Account       auth.Account
	Model         *mapping.ModelStruct
	UsernameField *mapping.StructField
	PasswordField *mapping.StructField
}

// HandleCheckUsername implements CheckUsernameHandler interface.
func (d *DefaultHandler) HandleCheckUsername(ctx context.Context, db database.DB, account auth.Account) error {
	cnt, err := db.QueryCtx(ctx, d.Model).
		Filter(filter.New(d.UsernameField, filter.OpEqual, account.GetUsername())).
		Count()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.WrapDet(auth.ErrAccountAlreadyExists, "account already exists").
			WithDetail("An account with provided username already exists")
	}
	return nil
}

// HandleRegisterAccount implements InsertAccountHandler interface.
func (d *DefaultHandler) HandleRegisterAccount(ctx context.Context, db database.DB, options *RegisterAccountOptions) error {
	return db.Insert(ctx, d.Model, options.Account)
}

// HandleLoginAccountRefresh implements LoginAccountRefreshHandler.
func (d *DefaultHandler) HandleLoginAccountRefresh(ctx context.Context, db database.DB, options *LoginOptions) error {
	return db.QueryCtx(ctx, d.Model, options.Account).Refresh()
}

// Initialize implements core.Initializer.
func (d *DefaultHandler) Initialize(c *controller.Controller) error {
	// Find the username field.
	var err error
	d.Model, err = c.ModelStruct(d.Account)
	if err != nil {
		return err
	}
	var ok bool
	d.UsernameField, ok = d.Model.FieldByName(d.Account.UsernameField())
	if !ok {
		return errors.Wrapf(auth.ErrInitialization, "provided invalid account model - no username field: '%s' found in the model: %s", d.Account.UsernameField(), d.Model)
	}
	d.PasswordField, ok = d.Model.FieldByName(d.Account.PasswordHashField())
	if !ok {
		return errors.Wrapf(auth.ErrInitialization, "provided invalid account model - no password hash field: '%s' found in the model: %s", d.Account.PasswordHashField(), d.Model)
	}
	return nil
}
