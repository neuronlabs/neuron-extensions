package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/migrate"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/mapping"
)

func testingController(t *testing.T, integration bool, models ...mapping.Model) *controller.Controller {
	t.Helper()

	c := controller.NewDefault()

	cfg := config.Service{DriverName: "postgres"}

	e, ok := os.LookupEnv("POSTGRES_TESTING")
	if integration && !ok {
		t.Skip("no 'POSTGRES_TESTING' environment variable defined")
	}
	cfg.Connection.RawURL = e

	err := c.RegisterService("postgres-testing", &cfg)
	require.NoError(t, err)
	if integration {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		err = c.DialAll(ctx)
		require.NoError(t, err)
	}
	require.NoError(t, c.RegisterModels(models...))
	if integration {
		require.NoError(t, c.MapModelsRepositories(models...))
		require.NoError(t, c.MigrateModels(context.Background(), models...))
	} else {
		require.NoError(t, migrate.PrepareModels(c.ModelMap.Models()...))
	}
	return c
}

func testingRepository(c *controller.Controller) *Postgres {
	return c.Services["postgres-testing"].(*Postgres)
}
