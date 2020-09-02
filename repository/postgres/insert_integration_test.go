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

func TestInsertSingleModel(t *testing.T) {
	db := testingDB(t, true, testModels...)
	p := testingRepository(db)

	ctx := context.Background()
	mStruct, err := db.ModelMap().ModelStruct(&tests.SimpleModel{})
	require.NoError(t, err)

	defer func() {
		_ = internal.DropTables(ctx, p.ConnPool, mStruct.DatabaseName, mStruct.DatabaseSchemaName)
		mStruct, err = db.ModelMap().ModelStruct(&tests.ForeignKeyModel{})
		require.NoError(t, err)
		_ = internal.DropTables(ctx, p.ConnPool, mStruct.DatabaseName, mStruct.DatabaseSchemaName)
	}()

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

	t.Run("ForeignKeyNotNull", func(t *testing.T) {
		model := &tests.ForeignKeyModel{}
		mStruct, err := db.ModelMap().ModelStruct(model)
		require.NoError(t, err)

		err = db.Query(mStruct, model).Select(mStruct.MustFieldByName("ForeignKey")).Insert()
		require.NoError(t, err)
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
		err = db.Query(mStruct, model1).
			Select(mStruct.MustFieldByName("Attr")).
			Insert()
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
			assert.True(t, errors.Is(err, query.ErrViolationUnique))
		}
	})
}
