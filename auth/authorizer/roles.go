package authorizer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/neuronlabs/neuron-extensions/auth/account"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/errors"
)

//go:generate neurogns models methods --format=goimports --single-file .
//go:generate neurogns collections --format=goimports  --single-file .

// Role is a simple role model for the RBAC authorization. It contains a many2many relation to Accounts.
type Role struct {
	ID uint
	// Timestamps
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
	Hierarchy int
	// Attributes
	Name        string `db:";unique_index"`
	Description string
	// Many2Many relations
	Accounts []*AccountRoles
	Scopes   []*AuthorizeScope `neuron:"many2many=RoleScopes;foreign=_,ScopeID"`
}

// RoleName implements auth.Role.
func (r *Role) RoleName() string {
	return r.Name
}

// HierarchyValue implements authorization.HierarchicalRole.
func (r *Role) HierarchyValue() int {
	return r.Hierarchy
}

// AccountRoles is the join model for the account roles many-to-many relationship.
type AccountRoles struct {
	ID uuid.UUID
	// Timestamps
	CreatedAt time.Time
	// Relations
	Role   *Role
	RoleID uint
	// AccountID is account foreign key converted to string.
	AccountID string `neuron:"type=foreign" db:"unique,index"`
}

// BeforeInsert implements database.BeforeInserter interface.
func (a *AccountRoles) BeforeInsert(context.Context, database.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// FindRoles implements authorization.Roler interface.
func (a *Authorizer) FindRoles(ctx context.Context, options ...auth.ListRoleOption) ([]auth.Role, error) {
	o := &auth.ListRoleOptions{}
	q := account.NRN_Roles.QueryCtx(ctx, a.db)
	for _, option := range options {
		option(o)
	}
	if o.Account != nil {
		q = q.Where("Accounts.AccountID IN", o.Account.GetPrimaryKeyValue())
	}
	if o.Limit > 0 {
		q.Limit(int64(o.Limit))
	}
	if o.Offset > 0 {
		q.Offset(int64(o.Offset))
	}
	roles, err := q.Find()
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, nil
	}

	roleInterfaces := make([]auth.Role, len(roles))
	for i := range roles {
		roleInterfaces[i] = roles[i]
	}
	return roleInterfaces, nil
}

// GrantRole implements authorization.Roler interface.
func (a *Authorizer) GrantRole(ctx context.Context, account auth.Account, role auth.Role) error {
	r, err := a.getRole(ctx, a.db, role)
	if err != nil {
		return err
	}
	if account.IsPrimaryKeyZero() {
		return errors.Wrap(auth.ErrAccountNotFound, "provided account has zero value primary key")
	}

	accountID, err := account.GetPrimaryKeyStringValue()
	if err != nil {
		return errors.Wrapf(auth.ErrAccountNotValid, "getting primary key value failed: %v", err)
	}

	cnt, err := NRN_AccountRoles.QueryCtx(ctx, a.db).Where("AccountID = ?", accountID).Count()
	if err != nil {
		return err
	}

	if cnt > 0 {
		return errors.Wrapf(auth.ErrRoleAlreadyGranted, "account already has role: %s", r.Name)
	}

	// Insert account roles.
	accRole := &AccountRoles{RoleID: r.ID, AccountID: accountID}
	if err = NRN_AccountRoles.Insert(ctx, a.db, accRole); err != nil {
		return err
	}
	return nil
}

// RevokeRole implements authorization.Roler interface.
func (a *Authorizer) RevokeRole(ctx context.Context, account auth.Account, role auth.Role) error {
	r, err := a.getRole(ctx, a.db, role)
	if err != nil {
		return err
	}
	if account.IsPrimaryKeyZero() {
		return errors.Wrap(auth.ErrAccountNotFound, "provided account has zero value primary key")
	}

	accountID, err := account.GetPrimaryKeyStringValue()
	if err != nil {
		return errors.Wrapf(auth.ErrAccountNotValid, "getting primary key value failed: %v", err)
	}
	_, err = NRN_AccountRoles.QueryCtx(ctx, a.db).
		Where("RoleID = ?", r.ID).
		Where("AccountID = ?", accountID).
		Delete()
	if err != nil {
		return err
	}
	return nil
}

// ClearRoles implements authorization.Roler interface.
func (a *Authorizer) ClearRoles(ctx context.Context, account auth.Account) error {
	deleted, err := NRN_AccountRoles.QueryCtx(ctx, a.db).Where("AccountID = ?").Delete()
	if err != nil {
		return err
	}
	log.Debugf("Cleared: '%d' roles for the account: '%v'", deleted, account.GetPrimaryKeyValue())
	return nil
}
