package authorizer

import (
	"context"
	"time"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query"
)

// AuthorizeScope is a scope used for the authorization process.
type AuthorizeScope struct {
	ID          uint
	CreatedAt   time.Time
	DeletedAt   *time.Time
	Name        string  `db:",index=unique"`
	Description string  `json:"omitempty"`
	Roles       []*Role `neuron:"many2many=RoleScopes;foreign=ScopeID"`
}

// ScopeName implements auth.Scope
func (a *AuthorizeScope) ScopeName() string {
	return a.Name
}

// RoleScopes is a join model for the mapping roles to scopes.
type RoleScopes struct {
	ID        uint64
	CreatedAt time.Time
	DeletedAt *time.Time
	Scope     *AuthorizeScope
	ScopeID   uint `db:";unique_index=idx_nrn_role_scopes_unique"`
	Role      *Role
	RoleID    uint `db:";unique_index=idx_nrn_role_scopes_unique"`
}

// ListRoleScopes lists the scopes for provided options.
func (a *Authorizer) ListRoleScopes(ctx context.Context, options ...auth.ListScopeOption) ([]auth.Scope, error) {
	o := &auth.ListScopeOptions{}
	for _, option := range options {
		option(o)
	}

	q := NRN_AuthorizeScopes.QueryCtx(ctx, a.db)
	if o.Limit > 0 {
		q.Limit(int64(o.Limit))
	}
	if o.Offset > 0 {
		q.Offset(int64(o.Offset))
	}
	if o.Role != nil {
		r, err := a.getRole(ctx, a.db, o.Role)
		if err != nil {
			return nil, err
		}
		q.Where("Roles.ID = ?", r.ID)
	}
	scopes, err := q.Find()
	if err != nil {
		return nil, err
	}
	result := make([]auth.Scope, len(scopes))
	for i, s := range scopes {
		result[i] = s
	}
	return result, nil
}

// ClearRoleScopes clears the scopes for provided roles/accounts.
func (a *Authorizer) ClearRoleScopes(ctx context.Context, roles ...auth.Role) error {
	if len(roles) == 0 {
		return errors.Wrap(auth.ErrInvalidRole, "provided no roles to clear")
	}
	err := database.RunInTransaction(ctx, a.db, nil, func(db database.DB) error {
		var roleIDs []interface{}
		for _, role := range roles {
			r, err := a.getRole(ctx, db, role)
			if err != nil {
				return err
			}
			roleIDs = append(roleIDs, r.ID)
		}
		deleted, err := NRN_RoleScopes.QueryCtx(ctx, db).
			Where("RoleID IN ?", roleIDs...).
			Delete()
		if err != nil {
			return err
		}
		log.Debugf("Deleted: %d role scopes", deleted)
		return nil
	})
	return err
}

// GrantRoleScope grants roles/accounts access for given scope.
func (a *Authorizer) GrantRoleScope(ctx context.Context, role auth.Role, scope auth.Scope) error {
	if role == nil {
		return errors.Wrap(auth.ErrInvalidRole, "provided invalid query")
	}
	if scope == nil {
		return errors.Wrap(auth.ErrAuthorizationScope, "provided no authorization scopes")
	}

	err := database.RunInTransaction(ctx, a.db, nil, func(db database.DB) error {
		r, err := a.getRole(ctx, db, role)
		if err != nil {
			return err
		}
		s, err := a.getScope(ctx, db, scope)
		if err != nil {
			return err
		}

		cnt, err := NRN_RoleScopes.QueryCtx(ctx, db).
			Where("RoleID = ?", r.ID).
			Where("ScopeID = ?", s.ID).
			Count()
		if cnt != 0 {
			return errors.Wrap(auth.ErrAuthorizationScope, "role already mapped to authorization scope")
		}
		rs := &RoleScopes{
			ScopeID: s.ID,
			RoleID:  r.ID,
		}
		return NRN_RoleScopes.Insert(ctx, db, rs)
	})
	return err
}

// RevokeRoleScope revokes the roles/accounts access for given scope.
func (a *Authorizer) RevokeRoleScope(ctx context.Context, role auth.Role, scope auth.Scope) error {
	if role == nil {
		return errors.Wrap(auth.ErrInvalidRole, "provided invalid query")
	}
	if scope == nil {
		return errors.Wrap(auth.ErrAuthorizationScope, "provided no authorization scopes")
	}

	err := database.RunInTransaction(ctx, a.db, nil, func(db database.DB) error {
		r, err := a.getRole(ctx, db, role)
		if err != nil {
			return err
		}
		s, err := a.getScope(ctx, db, scope)
		if err != nil {
			return err
		}

		// Delete all role scopes.
		cnt, err := NRN_RoleScopes.QueryCtx(ctx, db).
			Where("RoleID = ?", r.ID).
			Where("ScopeID = ?", s.ID).
			Delete()
		if cnt == 0 {
			return errors.Wrap(query.ErrNoResult, "nothing to revoke")
		}
		return nil
	})
	return err
}
