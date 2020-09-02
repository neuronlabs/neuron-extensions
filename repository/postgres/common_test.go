package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

func testingDB(t *testing.T, integration bool, models ...mapping.Model) *database.Database {
	t.Helper()

	mm := mapping.New(
		mapping.WithDefaultNotNull,
		mapping.WithDefaultDatabaseSchema("public"),
	)

	e, ok := os.LookupEnv("POSTGRES_TESTING")
	if integration && !ok {
		t.Skip("no 'POSTGRES_TESTING' environment variable defined")
	}
	var options []repository.Option
	if e != "" {
		options = append(options, repository.WithURI(e))
	}
	p := New(options...)
	databaseOptions := []database.Option{
		database.WithDefaultRepository(p),
		database.WithModelMap(mm),
	}
	if integration {
		databaseOptions = append(databaseOptions, database.WithMigrateModels(models...))
	}
	db, err := database.New(databaseOptions...)
	require.NoError(t, err)

	if integration {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		err = db.Dial(ctx)
		require.NoError(t, err)
	}
	return db
}

func testingRepository(db database.DB) *Postgres {
	repo, ok := db.(database.DefaultRepositoryGetter).GetDefaultRepository()
	if !ok {
		panic("no default repository found")
	}
	return repo.(*Postgres)
}
