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
	if cfg.RawURL != "" {
		cfgDB, err = pgxpool.ParseConfig(cfg.RawURL)
		if err != nil {
			return cfgDB, err
		}
	} else {
		connConfig := pgconn.Config{
			User:      cfg.Username,
			Password:  cfg.Password,
			Database:  cfg.DBName,
			Host:      cfg.Host,
			Port:      uint16(cfg.Port),
			TLSConfig: cfg.TLS,
		}
		if connConfig.Port == 0 {
			connConfig.Port = 5432
		}
		cfgDB = &pgxpool.Config{ConnConfig: &pgx.ConnConfig{Config: connConfig}}
	}
	return cfgDB, nil
}

// TestingConfig gets postgres config from the POSTGRES_TESTING environment variable.
func TestingConfig(t *testing.T) *pgxpool.Config {
	pg, ok := os.LookupEnv("POSTGRES_TESTING")
	if !ok {
		t.Skip("POSTGRES_TESTING environment variable not defined")
	}

	cfg, err := pgxpool.ParseConfig(pg)
	require.NoError(t, err)
	return cfg
}
