package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/repository"
	"github.com/neuronlabs/neuron/service"

	postgresErrors "github.com/neuronlabs/neuron-extensions/repository/postgres/errors"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/log"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/migrate"
)

// FactoryName defines the name of the factory.
const FactoryName = "postgres"

var (
	// Compile time check if Postgres implements repository.Repository interface.
	_ repository.Repository = &Postgres{}
	// compile time check for the service.Service interface.
	_ service.Service = &Postgres{}
	// compile time check for the service.Migrator interface.
	_ service.Migrator = &Postgres{}
)

// Postgres is the neuron repository that allows to query postgres databases.
// It allows to query PostgreSQL based databases using github.com/jackc/pgx driver.
// The repository implements:
//	- query.FullRepository
// 	- repository.Repository
//	- repository.Migrator
// The repository allows to share single transaction per multiple models - if all are registered within single database.
type Postgres struct {
	// id is the unique identification number of given repository instance.
	id                 uuid.UUID
	ConnPool           *pgxpool.Pool
	Config             *pgxpool.Config
	AutoSelectNotNulls bool
	// PostgresVersion is the numerical version of the postgres server.
	PostgresVersion int

	coreConfig   *config.Service
	keywords     map[string]migrate.KeyWordType
	Transactions map[uuid.UUID]pgx.Tx
}

// New creates new postgres instance.
func New() *Postgres {
	return &Postgres{
		id:                 uuid.New(),
		AutoSelectNotNulls: true,
		keywords:           map[string]migrate.KeyWordType{},
		Transactions:       map[uuid.UUID]pgx.Tx{},
	}
}

// ID returns unique repository id.
func (p *Postgres) ID() string {
	return p.id.String()
}

// Close closes given repository connections.
func (p *Postgres) Close(ctx context.Context) (err error) {
	p.ConnPool.Close()
	return nil
}

// Dial implements repository.Postgres interface. Creates a new Connection Pool for given repository.
func (p *Postgres) Dial(ctx context.Context) (err error) {
	if err = p.establishConnection(ctx); err != nil {
		return err
	}

	// Read postgres version.
	p.PostgresVersion, err = migrate.GetVersion(ctx, p.ConnPool)
	if err != nil {
		return err
	}

	// Get and store keywords for current postgres version.
	p.keywords, err = migrate.GetKeyWords(ctx, p.ConnPool, p.PostgresVersion)
	if err != nil {
		log.Errorf("Getting keywords for the postgres version: '%d' failed: %v", p.PostgresVersion, err)
		return err
	}
	return nil
}

// FactoryName returns the name of the factory for this Postgres.
// Implements repository.Repository interface.
func (p *Postgres) FactoryName() string {
	return FactoryName
}

// MigrateModels implements repository.Migrator interface.
// The method creates models tables if not exists and updates the columns per given model fields.
func (p *Postgres) MigrateModels(ctx context.Context, models ...*mapping.ModelStruct) error {
	if p.ConnPool == nil {
		return errors.Newf(service.ClassConnection, "no connection established")
	}
	for _, model := range models {
		if err := migrate.AutoMigrateModel(ctx, p.ConnPool, model); err != nil {
			return err
		}
	}
	return nil
}

// HealthCheck implements repository.Repository interface.
// It creates basic queries that checks if the connection is alive and returns given health response.
// The health response contains also notes with postgres version.
func (p *Postgres) HealthCheck(ctx context.Context) (*service.HealthResponse, error) {
	if p.ConnPool == nil {
		// if no pool is defined than no Dial method was done.
		return nil, errors.Newf(service.ClassConnection, "no connection established")
	}
	var temp string
	if err := p.ConnPool.QueryRow(ctx, "SELECT 1").Scan(&temp); err != nil {
		return &service.HealthResponse{
			Status: service.StatusFail,
			Output: err.Error(),
		}, nil
	}

	if err := p.ConnPool.QueryRow(ctx, "SELECT VERSION()").Scan(&temp); err != nil {
		return &service.HealthResponse{
			Status: service.StatusFail,
			Output: err.Error(),
		}, nil
	}
	// the repository is healthy.
	return &service.HealthResponse{
		Status: service.StatusPass,
		Notes:  []string{temp},
	}, nil
}

// RegisterModels implements repository.Repository interface.
func (p *Postgres) RegisterModels(models ...*mapping.ModelStruct) error {
	for _, model := range models {
		if err := migrate.PrepareModel(model); err != nil {
			return err
		}
	}
	return nil
}

/**

Private Methods

*/

// establishConnection Creates new database connection based on te provided DBConfig.
func (p *Postgres) establishConnection(ctx context.Context) (err error) {
	p.ConnPool, err = pgxpool.ConnectConfig(ctx, p.Config)
	if err != nil {
		return errors.NewDetf(service.ClassConnection, "cannot open database connection: %s", err.Error())
	}
	conn, err := p.ConnPool.Acquire(ctx)
	if err != nil {
		return errors.NewDetf(service.ClassConnection, "cannot open database connection: %s", err.Error())
	}
	if err = conn.Conn().Ping(ctx); err != nil {
		return errors.NewDet(service.ClassConnection, "cannot establish database connection for pq repository")
	}
	return nil
}

func (p *Postgres) errorClass(err error) errors.Class {
	mapped, ok := postgresErrors.Get(err)
	if ok {
		return mapped
	}
	return postgresErrors.ClassUnmappedError
}

func (p *Postgres) connection(s *query.Scope) internal.Connection {
	if tx := s.Transaction; tx != nil {
		return p.Transactions[tx.ID]
	}
	return p.ConnPool
}
