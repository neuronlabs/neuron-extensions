// +build integrate

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/controller"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/tests"
)

func init() {
}

var (
	tn                 = time.Now()
	testModelInstances = []*tests.Model{
		{ID: 1, AttrString: "some", Int: 2, CreatedAt: tn},
		{ID: 2, AttrString: "some", Int: 5, CreatedAt: tn},
		{ID: 3, AttrString: "some", Int: 1, CreatedAt: tn},
		{ID: 4, AttrString: "some", Int: 2, CreatedAt: tn},
	}
)

func prepareIntegrateRepository(t testing.TB) (*controller.Controller, *pgxpool.Pool) {
	cfg := internal.TestingConfig(t)
	t.Helper()

	// Create new controller.
	c := controller.NewDefault()
	// Register Service.
	err := c.RegisterService("postgres", cfg)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, c.DialAll(ctx))

	// register models
	require.NoError(t, c.RegisterModels(&tests.Model{}))
	require.NoError(t, c.MapModelsRepositories(&tests.Model{}))
	require.NoError(t, c.MigrateModels(ctx, &tests.Model{}))

	repo, err := c.GetRepository(&tests.Model{})
	require.NoError(t, err)

	pg, ok := repo.(*Postgres)
	require.True(t, ok)

	// fill the test model
	fillTestModelTable(t, pg.ConnPool)

	return c, pg.ConnPool
}

func fillTestModelTable(t testing.TB, db *pgxpool.Pool) {
	t.Helper()

	strPtr := "value"
	testModelInstances[0].StringPtr = &strPtr

	for _, model := range testModelInstances {
		_, err := db.Exec(context.Background(), `INSERT INTO test_models (attr_string, string_ptr, int, created_at) VALUES($1,$2,$3,$4);`, model.AttrString, model.StringPtr, model.Int, model.CreatedAt)
		require.NoError(t, err)
	}

}

func deleteTestModelTable(t testing.TB, db *pgxpool.Pool) {
	t.Helper()

	_, err := db.Exec(context.Background(), "DROP TABLE test_models;")
	require.NoError(t, err)
}
