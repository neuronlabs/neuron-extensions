package account

import (
	"context"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query"
)

var (
	_ auth.Authorizer     = &Authorizer{}
	_ auth.RoleAuthorizer = &Authorizer{}
)

// Authorizer is an implementation of the auth.Authorizer interface.
// It also implements full auth.RoleAuthorizer.
type Authorizer struct {
	c  *controller.Controller
	DB database.DB
}

// Initialize implements core.Initializer interface.
func (a *Authorizer) Initialize(c *controller.Controller) error {
	a.c = c
	a.DB = database.New(c)
	return nil
}

func (a *Authorizer) SetRoles(ctx context.Context, db database.DB, accountID interface{}, roles ...string) error {
	if len(roles) == 0 {
		return errors.NewDetf(query.ClassInvalidInput, "provided no roles to set")
	}
	acc, err := NRN_Accounts.QueryCtx(ctx, db).Where("ID = ?", accountID).Get()
	if err != nil {
		return err
	}
	roleInterfaces := make([]interface{}, len(roles))
	for i := range roles {
		roleInterfaces[i] = roles[i]
	}
	roleModels, err := NRN_Roles.QueryCtx(ctx, db).
		Where("Name IN", roleInterfaces...).
		Find()
	if err != nil {
		return err
	}
	if len(roleModels) != len(roles) {
		return errors.NewDetf(query.ClassInvalidInput, "one of provided roles doesn't exists: %v", roles)
	}

	return NRN_Accounts.SetRoles(ctx, db, acc, roleModels...)
}

// CreateRole creates a new 'role' with optional 'description'.
func (a *Authorizer) CreateRole(ctx context.Context, role, description string) (auth.Role, error) {
	r := &Role{Name: role, Description: description}
	if err := NRN_Roles.QueryCtx(ctx, a.DB, r).Insert(); err != nil {
		return nil, err
	}
	return r, nil
}

func (a *Authorizer) FindRoles(ctx context.Context, options ...auth.RoleFindOption) ([]auth.Role, error) {
	o := &auth.RoleFindOptions{}
	q := NRN_Roles.QueryCtx(ctx, a.DB)
	for _, option := range options {
		option(o)
	}
	if len(o.AccountIDs) > 0 {
		q = q.Where("Accounts.ID IN", o.AccountIDs...)
	}
	if len(o.Scopes) > 0 {

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

func (a *Authorizer) GetRoles(ctx context.Context, accountID interface{}) ([]auth.Role, error) {
	roles, err := NRN_Roles.QueryCtx(ctx, a.DB).
		Where("Accounts.ID = ?", accountID).
		Find()
	if err != nil {
		return nil, err
	}
	roleInterfaces := make([]auth.Role, len(roles))
	for i := range roles {
		roleInterfaces[i] = roles[i]
	}
	return roleInterfaces, nil
}

// DeleteRole implements auth.RoleAuthorizer interface.
func (a *Authorizer) DeleteRole(ctx context.Context, roleName string) error {
	role, err := NRN_Roles.QueryCtx(ctx, a.DB).Where("Name = ?", roleName).Get()
	if err != nil {
		return err
	}
	if _, err = NRN_Roles.QueryCtx(ctx, a.DB).Where("ID = ?", role.ID).Delete(); err != nil {
		return err
	}
	relationsCleared, err := NRN_Roles.ClearAccountsRelation(ctx, a.DB, role)
	if err != nil {
		return err
	}
	log.Debugf("Cleared: '%d' relations for role: '%d'", relationsCleared)
	return nil
}

func (a *Authorizer) GrantRole(ctx context.Context, db database.DB, role, scope string) error {
	s, err := NRN_AuthorizeScopes.QueryCtx(ctx, db).Where("Name = ?", scope).Get()
	if err != nil {
		return err
	}
	r, err := NRN_Roles.QueryCtx(ctx, db).Where("Name = ?", role).Get()
	if err != nil {
		return err
	}

	roleScope := &RoleScopes{
		ScopeID: s.ID,
		RoleID:  r.ID,
	}
	return NRN_RoleScopes.QueryCtx(ctx, db, roleScope).Insert()
}

func (a *Authorizer) RevokeRole(ctx context.Context, db database.DB, role, scope string) error {
	s, err := NRN_AuthorizeScopes.QueryCtx(ctx, db).Where("Name = ?", scope).Get()
	if err != nil {
		return err
	}
	r, err := NRN_Roles.QueryCtx(ctx, db).Where("Name = ?", role).Get()
	if err != nil {
		return err
	}

	_, err = NRN_RoleScopes.QueryCtx(ctx, db).
		Where("RoleID = ?", r.ID).
		Where("ScopeID = ?", s.ID).
		Delete()
	if err != nil {
		return err
	}
	return nil
}

// AddRole implements RoleAuthorizer interface.
func (a *Authorizer) AddRole(ctx context.Context, db database.DB, accountID interface{}, role string) error {
	r, err := NRN_Roles.QueryCtx(ctx, db).Where("Name = ?", role).Get()
	if err != nil {
		return err
	}

	acc, err := NRN_Accounts.QueryCtx(ctx, db).Where("ID = ?", accountID).Get()
	if err != nil {
		return err
	}

	accountRole := &AccountRoles{
		RoleID:    r.ID,
		AccountID: acc.ID,
	}
	return NRN_AccountRoles.QueryCtx(ctx, db, accountRole).Insert()
}

// ClearRoles implements RoleAuthorizer interface.
func (a *Authorizer) ClearRoles(ctx context.Context, db database.DB, accountID interface{}) error {
	acc, err := NRN_Accounts.QueryCtx(ctx, db).Where("ID = ?", accountID).Get()
	if err != nil {
		return err
	}
	rolesCleared, err := NRN_Accounts.ClearRolesRelation(ctx, db, acc)
	if err != nil {
		return err
	}
	log.Debugf("Cleared: '%d' roles for the account: '%s'", rolesCleared, acc.ID)
	return nil
}

// RemoveRole implements RoleAuthorizer interface.
func (a *Authorizer) RemoveRole(ctx context.Context, db database.DB, accountID interface{}, role string) error {
	r, err := NRN_Roles.QueryCtx(ctx, db).Where("Name = ?", role).Get()
	if err != nil {
		return err
	}

	deleted, err := NRN_AccountRoles.QueryCtx(ctx, db).
		Where("RoleID = ?", r.ID).
		Where("AccountID = ?", accountID).
		Delete()
	if err != nil {
		return err
	}
	if log.CurrentLevel().IsAllowed(log.LevelDebug) {
		if deleted > 0 {
			log.Debugf("Removed role: '%s' for account: %v", role, accountID)
		} else {
			log.Debugf("No role: '%s' found for the account: '%v'", role, accountID)
		}
	}
	return nil
}

// Authorize checks if the account with provided ID is allowed to use the 'resource'.
// Provided account must be authorized for ALL provided scopes.
func (a *Authorizer) Authorize(ctx context.Context, accountID interface{}, scopes ...string) error {
	if len(scopes) == 0 {
		return errors.NewDetf(auth.ClassScope, "provided no authorization scope")
	}

	roles, err := NRN_Roles.QueryCtx(ctx, a.DB).
		Where("Accounts.ID = ?", accountID).
		IncludeScopes("Name").
		Select("ID").
		Find()
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		return errors.NewDetf(auth.ClassForbidden, "not authorized")
	}

	mapped := map[string]bool{}
	for _, scope := range scopes {
		mapped[scope] = false
	}
	count := len(scopes)

checkLoop:
	for _, role := range roles {
		for _, scope := range role.Scopes {
			found, ok := mapped[scope.Name]
			if !ok || found {
				continue
			}
			mapped[scope.Name] = true
			count--
			if count == 0 {
				break checkLoop
			}
		}
	}
	if count == 0 {
		return nil
	}
	return errors.NewDetf(auth.ClassForbidden, "not authorized for all provided scopes")
}
