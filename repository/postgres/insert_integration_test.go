// +build integrate

package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/migrate"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/tests"
	"github.com/neuronlabs/neuron/db"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/query"
)

func TestInsertSingleModel(t *testing.T) {
	c := testingController(t, true, testModels...)
	p := testingRepository(c)

	ctx := context.Background()
	mStruct, err := c.ModelStruct(&tests.SimpleModel{})
	require.NoError(t, err)

	defer func() {
		table, err := migrate.ModelsTable(mStruct)
		require.NoError(t, err)
		_ = internal.DropTables(ctx, p.ConnPool, table.Name, table.Schema)
	}()

	// No results should return no error.
	db := db.New(c)

	newModel := func() *tests.SimpleModel {
		return &tests.SimpleModel{
			Attr: "Something",
		}
	}
	// Insert two models.
	t.Run("AutoFieldset", func(t *testing.T) {
		model1 := newModel()
		err = db.Query(mStruct, model1).Insert()
		require.NoError(t, err)

		assert.NotZero(t, model1.ID)
	})

	t.Run("BatchModels", func(t *testing.T) {
		model1 := newModel()
		model2 := newModel()
		err = db.Query(mStruct, model1, model2).Insert()
		require.NoError(t, err)

		assert.NotZero(t, model1.ID)
		assert.NotZero(t, model2.ID)

		assert.NotEqual(t, model1.ID, model2.ID)
	})

	t.Run("WithFieldset", func(t *testing.T) {
		model1 := newModel()
		model1.Attr = "something"
		err = db.Query(mStruct, model1).Select(mStruct.MustFieldByName("Attr")).Insert()
		require.NoError(t, err)

		assert.NotZero(t, model1.ID)
	})

	t.Run("WithID", func(t *testing.T) {
		model1 := newModel()
		model1.ID = 1e8
		err = db.Query(mStruct, model1).Insert()
		require.NoError(t, err)

		assert.NotZero(t, model1.ID)
		err = db.Query(mStruct, model1).Insert()
		if assert.Error(t, err) {
			assert.True(t, errors.IsClass(err, query.ClassViolationUnique), err)
		}
	})
}
