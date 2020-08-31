package tests

import (
	"time"

	"github.com/google/uuid"

	"github.com/neuronlabs/neuron-extensions/neurogonesis/internal/tests/external"
)

//go:generate neurogonesis models methods --format=goimports --single-file .
//go:generate neurogonesis collections  --single-file --format=goimports -o collections .

// User is testing model.
type User struct {
	ID            uuid.UUID
	CreatedAt     time.Time
	DeletedAt     *time.Time
	Name          *string
	Age           int
	IntArray      []int
	Bytes         []byte
	PtrBytes      *[]byte
	Wrapped       external.Int
	PtrWrapped    *external.Int
	External      *external.Model
	ExternalID    int
	FavoriteCar   Car
	FavoriteCarID string     `db:";notnull"`
	SisterID      *uuid.UUID `db:";notnull"`
	Cars          []*Car
	Sons          []*User   `neuron:"foreign=ParentID"`
	ParentID      uuid.UUID `neuron:"type=foreign"`
	Sister        *User
}

// Car is the test model for generator.
type Car struct {
	ID     string
	UserID uuid.UUID
	Plates string
}

// NeuronCollectionName implements mapping.Model.
func (c *Car) NeuronCollectionName() string {
	return "custom_cars"
}
