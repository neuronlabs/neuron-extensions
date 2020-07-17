package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/mapping"
)

func testingController(t *testing.T, dial bool, models ...mapping.Model) *controller.Controller {
	t.Helper()

	c := controller.NewDefault()

	cfg := config.Service{DriverName: "postgres"}

	e, ok := os.LookupEnv("POSTGRES_TESTING")
	if dial && !ok {
		t.Skip("no 'POSTGRES_TESTING' environment variable defined")
	}
	cfg.Connection.RawURL = e

	err := c.RegisterService("postgres-testing", &cfg)
	require.NoError(t, err)
	if dial {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		err = c.DialAll(ctx)
		require.NoError(t, err)
	}
	require.NoError(t, c.RegisterModels(models...))
	return c
}

func testingRepository(c *controller.Controller) *Postgres {
	return c.Services["postgres-testing"].(*Postgres)
}
