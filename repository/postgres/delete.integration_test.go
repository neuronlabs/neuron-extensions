// +build integrate

package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/mapping"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/tests"
)

// TestIntegrationDelete integration tests for the deleteQuery processes.
func TestIntegrationDelete(t *testing.T) {
	db := testingDB(t, true, testModels...)
	p := testingRepository(db)

	ctx := context.Background()

	mStruct, err := db.ModelMap().ModelStruct(&tests.SimpleModel{})
	require.NoError(t, err)

	defer func() {
		_ = internal.DropTables(ctx, p.ConnPool, mStruct.DatabaseName, mStruct.DatabaseSchemaName)
	}()

	newModel := func() *tests.SimpleModel {
		return &tests.SimpleModel{
			Attr: "Something",
		}
	}
	t.Run("WithFilter", func(t *testing.T) {
		model := newModel()
		model2 := newModel()
		models := []mapping.Model{model, model2}
		// Insert models.
		err = db.Query(mStruct, models...).Insert()
		require.NoError(t, err)

		assert.Len(t, models, 2)

		affected, err := db.Query(mStruct).Where("ID IN", model.ID, model2.ID).Delete()
		require.NoError(t, err)

		assert.Equal(t, int64(2), affected)
	})

	t.Run("Models", func(t *testing.T) {
		model := newModel()
		model2 := newModel()
		// Insert models.
		err = db.Query(mStruct, model, model2).Insert()
		require.NoError(t, err)

		affected, err := db.Query(mStruct, model, model2).Delete()
		require.NoError(t, err)

		assert.Equal(t, int64(2), affected)
	})
}

func TestSoftDelete(t *testing.T) {
	db := testingDB(t, true, testModels...)
	p := testingRepository(db)

	ctx := context.Background()

	mStruct, err := db.ModelMap().ModelStruct(&tests.Model{})
	require.NoError(t, err)

	defer func() {
		_ = internal.DropTables(ctx, p.ConnPool, mStruct.DatabaseName, mStruct.DatabaseSchemaName)
	}()

	newModel := func() *tests.Model {
		return &tests.Model{
			AttrString: "Something",
		}
	}
	model := newModel()
	model2 := newModel()
	models := []mapping.Model{model, model2}
	// Insert models.
	err = db.Query(mStruct, models...).Insert()
	require.NoError(t, err)

	affected, err := db.Query(mStruct, model).Delete()
	require.NoError(t, err)

	assert.Equal(t, int64(1), affected)

	res, err := db.Query(mStruct).Where("ID = ", model.ID).Find()
	require.NoError(t, err)

	assert.Len(t, res, 0)
}
