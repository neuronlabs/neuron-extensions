package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/service"
)

/**

Factory

*/

// compile time check for the Factory interface implementation.
var _ service.Factory = &Factory{}

// Factory is the repository.Factory mock implementation
type Factory struct {
	mock.Mock
}

// New creates new repository
// Implements repository.Factory method
func (f *Factory) New(model *config.Repository) (service.Service, error) {
	return &Repository{id: uuid.New()}, nil
}

// DriverName returns the factory repository name
// Implements repository.Repository
func (f *Factory) DriverName() string {
	return DriverName
}

// Close closes the factory
func (f *Factory) Close(ctx context.Context, done chan<- interface{}) {
	done <- struct{}{}
}
