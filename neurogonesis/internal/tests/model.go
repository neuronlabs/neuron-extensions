package tests

import (
	"time"

	"github.com/google/uuid"

	"github.com/neuronlabs/neuron-extensions/neurogonesis/internal/tests/external"
)

//go:generate neurogonesis models methods --format=goimports --single-file .
//go:generate neurogonesis collections  --single-file --format=goimports -o collections .
//go:generate neurogonesis collections package User usercollection
//go:generate neurogonesis collections package Car cardb

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
	CustomUUID    UUID
}

// Car is the test model for generator.
type Car struct {
	ID               string
	UserID           uuid.UUID
	Plates           string
	Directory        SqlEnum
	EnumField        Enum
	UintEnumField    UintEnum
	Models           external.Models
	NonPointerModels external.NonPointerModels
}

// NeuronCollectionName implements mapping.Model.
func (c *Car) NeuronCollectionName() string {
	return "custom_cars"
}

const Size int = 16

type UUID [Size]byte

type Enum int

const (
	_ Enum = iota
	FirstE
	SecondE
)

type UintEnum uint

func (u UintEnum) String() string {
	return "value"
}

func (u *UintEnum) Do() {

}

const (
	_ UintEnum = iota
	First
	Second
)
