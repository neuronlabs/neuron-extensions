// +build integrate

package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/tests"
)

// // TestIntegrationPatch integration tests for update method.
func TestUpdate(t *testing.T) {
	db := testingDB(t, true, &tests.SimpleModel{})
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

	// Insert two models.
	model1 := newModel()
	model2 := newModel()
	err = db.Query(mStruct, model1, model2).Insert()
	require.NoError(t, err)

	t.Run("Model", func(t *testing.T) {
		model := newModel()
		model.ID = model1.ID
		model.Attr = "Other"
		affected, err := db.Query(mStruct, model).Update()
		require.NoError(t, err)

		assert.Equal(t, int64(1), affected)

		models, err := db.Query(mStruct).Where("ID =", model1.ID).Find()
		require.NoError(t, err)
		if assert.Len(t, models, 1) {
			assert.Equal(t, "Other", models[0].(*tests.SimpleModel).Attr)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		model := newModel()
		model.ID = 1e8
		affected, err := db.Query(mStruct, model).Update()
		require.NoError(t, err)

		assert.Equal(t, int64(0), affected)
	})

	t.Run("Filters", func(t *testing.T) {
		model := newModel()
		model.Attr = "Something"
		affected, err := db.Query(mStruct, model).Select(mStruct.MustFieldByName("Attr")).Where("ID in", model1.ID, model2.ID).Update()
		require.NoError(t, err)

		assert.Equal(t, int64(2), affected)
	})
}

func TestUpdateForeign(t *testing.T) {
	db := testingDB(t, true, &tests.ForeignKeyModel{})
	p := testingRepository(db)

	ctx := context.Background()

	mStruct, err := db.ModelMap().ModelStruct(&tests.ForeignKeyModel{})
	require.NoError(t, err)

	defer func() {
		_ = internal.DropTables(ctx, p.ConnPool, mStruct.DatabaseName, mStruct.DatabaseSchemaName)
	}()

	model := &tests.ForeignKeyModel{}
	err = db.Query(mStruct, model).Insert()
	require.NoError(t, err)

	_, err = db.Query(mStruct, model).Select(mStruct.MustFieldByName("ForeignKey")).Update()
	require.NoError(t, err)
}
