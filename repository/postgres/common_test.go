package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/migrate"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

func testingController(t *testing.T, integration bool, models ...mapping.Model) *controller.Controller {
	t.Helper()

	c := controller.NewDefault()
	e, ok := os.LookupEnv("POSTGRES_TESTING")
	if integration && !ok {
		t.Skip("no 'POSTGRES_TESTING' environment variable defined")
	}
	var options []repository.Option
	if e != "" {
		options = append(options, repository.WithURI(e))
	}
	p := New(options...)

	err := c.SetDefaultRepository(p)
	require.NoError(t, err)

	if integration {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		err = c.DialAll(ctx)
		require.NoError(t, err)
	}

	require.NoError(t, c.RegisterModels(models...))
	if integration {
		require.NoError(t, c.MigrateModels(context.Background(), models...))
	} else {
		require.NoError(t, migrate.PrepareModels(c.ModelMap.Models()...))
	}
	return c
}

func testingRepository(c *controller.Controller) *Postgres {
	for _, p := range c.Repositories {
		return p.(*Postgres)
	}
	return nil
}
