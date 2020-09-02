// +build integrate

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/tests"
	"github.com/neuronlabs/neuron/query/filter"
)

func TestArrayIntegrationModel(t *testing.T) {
	db := testingDB(t, true, &tests.ArrayModel{})
	p := testingRepository(db)

	ctx := context.Background()
	mStruct, err := db.ModelMap().ModelStruct(&tests.ArrayModel{})
	require.NoError(t, err)

	defer func() {
		_ = internal.DropTables(ctx, p.ConnPool, mStruct.DatabaseName, mStruct.DatabaseSchemaName)
	}()

	model := &tests.ArrayModel{
		ID:          uuid.New(),
		SliceInt:    []int{1, 2, 3},
		SliceString: []string{"4", "5", "6"},
	}

	mStruct = db.ModelMap().MustModelStruct(model)
	err = db.QueryCtx(ctx, mStruct, model).Insert()
	require.NoError(t, err)

	fromDB, err := db.QueryCtx(ctx, mStruct).Filter(filter.New(mStruct.Primary(), filter.OpEqual, model.ID)).Get()
	require.NoError(t, err)

	fromDBA, ok := fromDB.(*tests.ArrayModel)
	require.True(t, ok)
	assert.True(t, fromDBA.ID == model.ID)
	assert.Len(t, fromDBA.SliceInt, 3)
	assert.Len(t, fromDBA.SliceString, 3)
}
