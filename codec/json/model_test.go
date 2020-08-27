package json

import (
	"time"
)

//go:generate neurogonesis models methods --format=goimports --single-file .

// User is a test model.
type User struct {
	ID           int
	Name         string `codec:"first_name"`
	CreatedAt    time.Time
	CreatedAtIso time.Time `codec:"iso8601"`
	Mother       *User     `codec:"omitempty"`
	MotherID     int       `codec:"omitempty"`
	Father       *User
	FatherID     int    `codec:"omitempty"`
	Pets         []*Pet `codec:"omitempty" neuron:"foreign=OwnerID"`
}

// Pet is a test relation model.
type Pet struct {
	ID      int
	Name    string
	Owner   *User `codec:"omitempty"`
	OwnerID int   `codec:"omitempty" neuron:"type=foreign"`
}
