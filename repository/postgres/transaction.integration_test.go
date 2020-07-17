// +build integrate

package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/internal"
	"github.com/neuronlabs/neuron-plugins/repository/postgres/migrate"
	"github.com/neuronlabs/neuron-plugins/repository/postgres/tests"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/query"
)

func TestTransactions(t *testing.T) {
	c := testingController(t, true, &tests.SimpleModel{})
	p := testingRepository(c)

	ctx := context.Background()

	err := c.MigrateModels(ctx, &tests.SimpleModel{})
	require.NoError(t, err)

	mStruct, err := c.ModelStruct(&tests.SimpleModel{})
	require.NoError(t, err)

	defer func() {
		table, err := migrate.ModelsTable(mStruct)
		require.NoError(t, err)
		_ = internal.DropTables(ctx, p.ConnPool, table.Name, table.Schema)
	}()

	qc := query.NewCreator(c)

	t.Run("Commit", func(t *testing.T) {
		// No results should return no error.
		tx := query.Begin(ctx, c, nil)

		model := &tests.SimpleModel{Attr: "Name"}
		err = tx.Query(mStruct, model).Insert()
		require.NoError(t, err)

		_, ok := p.Transactions[tx.Transaction.ID]
		assert.True(t, ok)

		assert.NotEqual(t, 0, model.ID)
		err = tx.Commit()
		require.NoError(t, err)
		_, ok = p.Transactions[tx.Transaction.ID]
		assert.False(t, ok)

		res, err := qc.Query(mStruct).Where("id =", model.ID).Get()
		require.NoError(t, err)

		assert.Equal(t, res.GetPrimaryKeyValue(), model.ID)
	})

	t.Run("Rollback", func(t *testing.T) {
		// No results should return no error.
		tx := query.Begin(ctx, c, nil)

		model := &tests.SimpleModel{Attr: "Name"}
		err = tx.Query(mStruct, model).Insert()
		require.NoError(t, err)

		_, ok := p.Transactions[tx.Transaction.ID]
		assert.True(t, ok)

		assert.NotEqual(t, 0, model.ID)

		err = tx.Rollback()
		require.NoError(t, err)

		_, ok = p.Transactions[tx.Transaction.ID]
		assert.False(t, ok)

		_, err := qc.Query(mStruct).Where("id =", model.ID).Get()
		require.Error(t, err)
		assert.True(t, errors.IsClass(err, query.ClassNoResult))
	})
}
