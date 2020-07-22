package internal

import (
	"os"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/config"
)

// RepositoryConfig gets the *ConfigDB based on the config.ModelConfig.
func RepositoryConfig(cfg *config.Service) (cfgDB *pgxpool.Config, err error) {
	if cfg.Connection.RawURL != "" {
		cfgDB, err = pgxpool.ParseConfig(cfg.Connection.RawURL)
		if err != nil {
			return cfgDB, err
		}
	} else {
		connConfig := pgconn.Config{
			User:      cfg.Connection.Username,
			Password:  cfg.Connection.Password,
			Database:  cfg.DBName,
			Host:      cfg.Connection.Host,
			Port:      uint16(cfg.Connection.Port),
			TLSConfig: cfg.Connection.TLS,
		}
		if connConfig.Port == 0 {
			connConfig.Port = 5432
		}
		cfgDB = &pgxpool.Config{ConnConfig: &pgx.ConnConfig{Config: connConfig}}
	}
	return cfgDB, nil
}

// TestingPostgresConfig gets postgres config from the POSTGRES_TESTING environment variable.
func TestingPostgresConfig(t testing.TB) *pgxpool.Config {
	pg, ok := os.LookupEnv("POSTGRES_TESTING")
	if !ok {
		t.Skip("POSTGRES_TESTING environment variable not defined")
	}

	cfg, err := pgxpool.ParseConfig(pg)
	require.NoError(t, err)
	return cfg
}

func TestingConfig(t testing.TB) *config.Service {
	pg, ok := os.LookupEnv("POSTGRES_TESTING")
	if !ok {
		t.Skip("POSTGRES_TESTING environment variable not defined")
	}
	return &config.Service{Connection: config.Connection{RawURL: pg}}
}
