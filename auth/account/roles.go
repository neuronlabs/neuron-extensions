package account

import (
	"time"

	"github.com/google/uuid"
)

// Role is a simple role model for the RBAC authorization. It contains a many2many relation to Accounts.
type Role struct {
	ID uint
	// Timestamps
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
	// Attributes
	Name        string `db:"index=unique"`
	Description string
	// Many2Many relations
	Accounts []*Account
	Scopes   []*AuthorizeScope `neuron:"many2many=RoleScopes;foreign=,ScopeID"`
}

// RoleName implements auth.Role.
func (r *Role) RoleName() string {
	return r.Name
}

// AccountRoles is the join model for the account roles many-to-many relationship.
type AccountRoles struct {
	ID uuid.UUID
	// Timestamps
	CreatedAt time.Time
	DeletedAt *time.Time
	// Join Model relations.
	Role      *Role
	RoleID    uint
	Account   *Account
	AccountID uuid.UUID
}
