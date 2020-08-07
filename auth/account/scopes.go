package account

import (
	"time"
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

// RoleScopes is a join model for the mapping roles to scopes.
type RoleScopes struct {
	ID        uint64
	CreatedAt time.Time
	DeletedAt *time.Time
	Scope     *AuthorizeScope
	ScopeID   uint `db:",index=unique_roles,unique"`
	Role      *Role
	RoleID    uint `db:",index=unique_roles,unique"`
}
