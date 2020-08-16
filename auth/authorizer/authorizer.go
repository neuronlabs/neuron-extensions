package authorizer

import (
	"context"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/repository"
)

var (
	_ auth.Verifier = &Authorizer{}
	_ auth.Roler    = &Authorizer{}
)

// Authorizer is an implementation of the auth.Verifier, auth.Roler, auth.Scoper interfaces.
// It gives the basic authorization functions with use of following models:
//	- Role
//	- ScopeRoles
//	- AccountRoles
//	- AuthorizationScope
// 	By default it migrates these models into default controllers repository.
// 	If specific repository is prepared for the authorizer models, it could be provided as an option to the `New` function.
type Authorizer struct {
	Options *Options

	c  *controller.Controller
	db database.DB
}

// Options are the Authorizer options.
type Options struct {
	Repository    repository.Repository
	MigrateModels bool
}

// Option is the function that sets the Options.
type Option func(o *Options)

// WithRepository sets the repository in the authorizer options.
// If no repository is defined than the default controllers repository would be taken.
func WithRepository(r repository.Repository) Option {
	return func(o *Options) {
		o.Repository = r
	}
}

// WithModelMigration migrates the models into provided repository.
func WithModelMigration() Option {
	return func(o *Options) {
		o.MigrateModels = true
	}
}

// New creates new Authorizer with provided creation options.
func New(options ...Option) *Authorizer {
	o := &Options{
		MigrateModels: true,
	}
	for _, option := range options {
		option(o)
	}
	return &Authorizer{
		Options: o,
	}
}

// Initialize implements core.Initializer interface.
func (a *Authorizer) Initialize(c *controller.Controller) error {
	a.c = c
	a.db = database.New(c)
	return nil
}

// Authorize checks if the account with provided ID is allowed to use the 'resource'.
// Provided account must be authorized for ALL provided scopes.
func (a *Authorizer) Verify(ctx context.Context, account auth.Account, options ...auth.VerifyOption) error {
	accountID, err := account.GetPrimaryKeyStringValue()
	if err != nil {
		return errors.Wrapf(auth.ErrAccountNotValid, "getting primary key value failed: %v", err)
	}

	o := &auth.VerifyOptions{}
	for _, option := range options {
		option(o)
	}

	// Prepare the query with provided accountID.
	q := NRN_AccountRoles.QueryCtx(ctx, a.db).
		Where("AccountID = ?", accountID)

	// Check allowed roles.
	if len(o.AllowedRoles) > 0 {
		names := make([]interface{}, len(o.AllowedRoles))
		for i, r := range o.AllowedRoles {
			names[i] = r.RoleName()
		}
		q.Where("Role.Name IN ?", names...)
	}

	// Check disallowed roles.
	if len(o.DisallowedRoles) > 0 {
		names := make([]interface{}, len(o.DisallowedRoles))
		for i, r := range o.DisallowedRoles {
			names[i] = r.RoleName()
		}
		q.Where("Role.Name NOT IN ?", names...)
	}

	// Scope queries.
	if len(o.Scopes) > 0 {
		names := make([]interface{}, len(o.Scopes))
		for i, s := range o.Scopes {
			names[i] = s.ScopeName()
		}
		scopes, err := NRN_AuthorizeScopes.QueryCtx(ctx, a.db).
			Where("Name IN ?", names...).
			IncludeRoles("ID").
			Find()
		if err != nil {
			return err
		}
		roleIDs := map[uint]struct{}{}
		for _, scope := range scopes {
			for _, r := range scope.Roles {
				roleIDs[r.ID] = struct{}{}
			}
		}
		var filterIDs []interface{}
		for id := range roleIDs {
			filterIDs = append(filterIDs, id)
		}
		q.Where("RoleID IN ?", filterIDs...)
	}

	count, err := q.Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	return errors.WrapDetf(auth.ErrForbidden, "not authorized")
}

func (a *Authorizer) getRole(ctx context.Context, db database.DB, role auth.Role) (*Role, error) {
	var (
		r   *Role
		ok  bool
		err error
	)
	if r, ok = role.(*Role); ok {
		if r.ID == 0 && r.Name == "" {
			return nil, errors.Wrap(auth.ErrInvalidRole, "provided invalid role")
		}
		if r.ID > 0 && r.Name == "" {
			if err = NRN_Roles.Refresh(ctx, db, r); err != nil {
				if errors.Is(err, query.ErrNoResult) {
					return nil, errors.Wrap(auth.ErrInvalidRole, "no such role")
				}
				return nil, err
			}
		} else if r.ID == 0 && r.Name != "" {
			r, err = NRN_Roles.QueryCtx(ctx, db).Where("Name = ?", r.Name).Get()
			if err != nil {
				if errors.Is(err, query.ErrNoResult) {
					return nil, errors.Wrap(auth.ErrInvalidRole, "no such role")
				}
				return nil, err
			}
		}
	} else {
		if role.RoleName() == "" {
			return nil, errors.Wrap(auth.ErrInvalidRole, "provided role without name")
		}
		r, err = NRN_Roles.QueryCtx(ctx, db).Where("Name = ?", role.RoleName()).Get()
		if err != nil {
			if errors.Is(err, query.ErrNoResult) {
				return nil, errors.Wrap(auth.ErrInvalidRole, "no such role")
			}
			return nil, err
		}
	}
	return r, nil
}

func (a *Authorizer) getScope(ctx context.Context, db database.DB, scope auth.Scope) (*AuthorizeScope, error) {
	var (
		s   *AuthorizeScope
		ok  bool
		err error
	)
	if s, ok = scope.(*AuthorizeScope); ok {
		if s.ID == 0 && s.Name == "" {
			return nil, errors.Wrap(auth.ErrAuthorizationScope, "provided invalid authorization scope")
		}
		if s.ID > 0 && s.Name == "" {
			if err = NRN_AuthorizeScopes.Refresh(ctx, db, s); err != nil {
				if errors.Is(err, query.ErrNoResult) {
					return nil, errors.Wrap(auth.ErrInvalidRole, "no such role")
				}
				return nil, err
			}
		} else if s.ID == 0 && s.Name != "" {
			s, err = NRN_AuthorizeScopes.QueryCtx(ctx, db).Where("Name = ?", s.Name).Get()
			if err != nil {
				if errors.Is(err, query.ErrNoResult) {
					return nil, errors.Wrap(auth.ErrInvalidRole, "no such role")
				}
				return nil, err
			}
		}
	} else {
		if s.ScopeName() == "" {
			return nil, errors.Wrap(auth.ErrAuthorizationScope, "provided scope without name")
		}
		s, err = NRN_AuthorizeScopes.
			QueryCtx(ctx, db).
			Where("Name = ?", scope.ScopeName()).
			Get()
		if err != nil {
			if errors.Is(err, query.ErrNoResult) {
				return nil, errors.Wrap(auth.ErrAuthorizationScope, "no such scope")
			}
			return nil, err
		}
	}
	return s, nil
}
