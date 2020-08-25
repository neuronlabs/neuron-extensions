package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

func testingController(t *testing.T, integration bool, models ...mapping.Model) *core.Controller {
	t.Helper()

	c := core.NewDefault()
	c.Options.DefaultNotNullFields = true
	c.ModelMap.Options.DefaultNotNull = true

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
	require.NoError(t, c.MapRepositoryModels(p, models...))
	require.NoError(t, c.SetUnmappedModelRepositories())
	require.NoError(t, c.RegisterRepositoryModels())
	if integration {
		require.NoError(t, c.MigrateModels(context.Background(), models...))
	}
	return c
}

func testingRepository(c *core.Controller) *Postgres {
	for _, p := range c.Repositories {
		return p.(*Postgres)
	}
	return nil
}
