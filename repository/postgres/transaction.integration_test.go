// +build integrate

package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/tests"
)

func TestTransactions(t *testing.T) {
	db := testingDB(t, true, testModels...)
	p := testingRepository(db)

	ctx := context.Background()

	mStruct, err := db.ModelMap().ModelStruct(&tests.SimpleModel{})
	require.NoError(t, err)

	defer func() {
		_ = internal.DropTables(ctx, p.ConnPool, mStruct.DatabaseName, mStruct.DatabaseSchemaName)
	}()

	t.Run("Commit", func(t *testing.T) {
		// No results should return no error.
		tx := db.Begin(ctx, nil)

		model := &tests.SimpleModel{Attr: "Name"}
		err = tx.Query(mStruct, model).Insert()
		require.NoError(t, err)

		_, ok := p.transactions[tx.Transaction.ID]
		assert.True(t, ok)

		assert.NotEqual(t, 0, model.ID)
		err = tx.Commit()
		require.NoError(t, err)
		_, ok = p.transactions[tx.Transaction.ID]
		assert.False(t, ok)

		res, err := db.Query(mStruct).Where("id =", model.ID).Get()
		require.NoError(t, err)

		assert.Equal(t, res.GetPrimaryKeyValue(), model.ID)
	})

	t.Run("Rollback", func(t *testing.T) {
		// No results should return no error.
		tx := db.Begin(ctx, nil)

		model := &tests.SimpleModel{Attr: "Name"}
		err = tx.Query(mStruct, model).Insert()
		require.NoError(t, err)

		_, ok := p.transactions[tx.Transaction.ID]
		assert.True(t, ok)

		assert.NotEqual(t, 0, model.ID)

		err = tx.Rollback()
		require.NoError(t, err)

		_, ok = p.transactions[tx.Transaction.ID]
		assert.False(t, ok)

		_, err := db.Query(mStruct).Where("id =", model.ID).Get()
		require.Error(t, err)
		assert.True(t, errors.Is(err, query.ErrNoResult))
	})
}

// TestSavepoints tests the savepoints for the database.
func TestSavepoints(t *testing.T) {
	db := testingDB(t, true, testModels...)
	p := testingRepository(db)

	ctx := context.Background()

	mStruct, err := db.ModelMap().ModelStruct(&tests.SimpleModel{})
	require.NoError(t, err)

	defer func() {
		_ = internal.DropTables(ctx, p.ConnPool, mStruct.DatabaseName, mStruct.DatabaseSchemaName)
	}()

	// No results should return no error.
	tx := db.Begin(ctx, nil)

	model := &tests.SimpleModel{Attr: "Name"}
	err = tx.Query(mStruct, model).Insert()
	require.NoError(t, err)

	err = tx.Savepoint("testing")
	require.NoError(t, err)

	model2 := &tests.SimpleModel{Attr: "Other"}
	err = tx.Query(mStruct, model2).Insert()
	require.NoError(t, err)

	err = tx.RollbackSavepoint("testing")
	require.NoError(t, err)

	cnt, err := tx.Query(mStruct).Count()
	require.NoError(t, err)
	assert.Equal(t, int64(1), cnt)

	err = tx.Rollback()
	require.NoError(t, err)
}
