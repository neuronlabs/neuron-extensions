// +build integrate

package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
	"github.com/neuronlabs/neuron-extensions/repository/postgres/tests"
	"github.com/neuronlabs/neuron/mapping"
)

var testModels = tests.Neuron_Models

// TestIntegrationCount does integration tests for the Count method.
func TestIntegrationCount(t *testing.T) {
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
			Int:        3,
		}
	}
	models := []mapping.Model{newModel(), newModel()}
	// Insert models.
	err = db.Query(mStruct, models...).Insert()
	require.NoError(t, err)

	assert.Len(t, models, 2)

	countAll, err := db.Query(mStruct).Count()
	require.NoError(t, err)

	assert.Equal(t, int64(2), countAll)
}
