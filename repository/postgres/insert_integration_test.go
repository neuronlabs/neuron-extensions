// +build integrate

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func TestInsertErrors(t *testing.T) {
	db := testingDB(t, true, testModels...)
	p := testingRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	mm := db.ModelMap()
	mStruct, err := mm.ModelStruct(&tests.SimpleModel{})
	require.NoError(t, err)

	fKMstruct, err := mm.ModelStruct(&tests.ForeignKeyModel{})
	require.NoError(t, err)

	defer func() {
		for _, model := range testModels {
			mStruct, err = mm.ModelStruct(model)
			require.NoError(t, err)
			_ = internal.DropTables(ctx, p.ConnPool, mStruct.DatabaseName, mStruct.DatabaseSchemaName)
		}
	}()

	_, err = p.ConnPool.Exec(ctx, fmt.Sprintf("ALTER TABLE %s.%s ADD CONSTRAINT simple_fk FOREIGN KEY (%s) REFERENCES %s.%s (%s)",
		fKMstruct.DatabaseSchemaName, fKMstruct.DatabaseName, fKMstruct.MustFieldByName("ForeignKey").DatabaseName,
		mStruct.DatabaseSchemaName, mStruct.DatabaseName, mStruct.Primary().DatabaseName))
	require.NoError(t, err)

	defer func() {
		_, err = p.ConnPool.Exec(ctx, fmt.Sprintf("ALTER TABLE %s.%s DROP CONSTRAINT simple_fk", fKMstruct.DatabaseSchemaName, fKMstruct.DatabaseName))
		require.NoError(t, err)
	}()

	t.Run("ViolationFK", func(t *testing.T) {
		t.Run("Bulk", func(t *testing.T) {
			err = db.Insert(ctx, fKMstruct, &tests.ForeignKeyModel{ForeignKey: 20}, &tests.ForeignKeyModel{ForeignKey: 21})
			if assert.Error(t, err) {
				assert.True(t, errors.Is(err, query.ErrViolationForeignKey))
			}
		})

		t.Run("Single", func(t *testing.T) {
			err = db.Insert(ctx, fKMstruct, &tests.ForeignKeyModel{ForeignKey: 20}, &tests.ForeignKeyModel{ForeignKey: 21})
			if assert.Error(t, err) {
				assert.True(t, errors.Is(err, query.ErrViolationForeignKey))
			}
		})
	})

	t.Run("Unique", func(t *testing.T) {
		simple := &tests.SimpleModel{}
		simple2 := &tests.SimpleModel{}
		simple3 := &tests.SimpleModel{}
		err = db.Insert(ctx, mStruct, simple, simple2, simple3)
		require.NoError(t, err)

		fk := &tests.ForeignKeyModel{ForeignKey: simple.ID}
		err = db.Insert(ctx, fKMstruct, fk)
		require.NoError(t, err)

		fk2 := &tests.ForeignKeyModel{ForeignKey: simple2.ID}
		err = db.Insert(ctx, fKMstruct, fk2)
		require.NoError(t, err)

		t.Run("Single", func(t *testing.T) {
			err = db.Insert(ctx, fKMstruct, fk2)
			if assert.Error(t, err) {
				assert.True(t, errors.Is(err, query.ErrViolationUnique))
			}
		})

		t.Run("Bulk", func(t *testing.T) {
			err = db.Insert(ctx, fKMstruct, &tests.ForeignKeyModel{ID: fk.ID, ForeignKey: simple3.ID}, &tests.ForeignKeyModel{ID: fk2.ID, ForeignKey: simple3.ID})
			if assert.Error(t, err) {
				assert.True(t, errors.Is(err, query.ErrViolationUnique))
			}
		})
	})
}
