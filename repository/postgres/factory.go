package postgres

import (
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/service"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/internal"
	"github.com/neuronlabs/neuron-plugins/repository/postgres/log"
)

func init() {
	if err := service.RegisterFactory(factory); err != nil {
		panic(err)
	}
}

var factory service.Factory = &Factory{}

// Factory is the pq.Postgres factory.
type Factory struct{}

// DriverName gets the Factory repository name.
// Implements repository.Postgres interface.
func (f *Factory) DriverName() string {
	return FactoryName
}

// New creates new PQ repository for the provided config 'cfg'.
func (f *Factory) New(cfg *config.Service) (service.Service, error) {
	repoConfig, err := internal.RepositoryConfig(cfg)
	if err != nil {
		log.Debugf("Getting postgres repository config failed: %v", err)
		return nil, err
	}
	r := New()
	r.Config = repoConfig
	r.coreConfig = cfg
	return r, nil
}
