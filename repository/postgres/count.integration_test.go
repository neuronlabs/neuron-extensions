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
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// TestIntegrationCount does integration tests for the Count method.
func TestIntegrationCount(t *testing.T) {
	c := testingController(t, true, &tests.Model{})
	p := testingRepository(c)

	ctx := context.Background()

	err := c.MigrateModels(ctx, &tests.Model{})
	require.NoError(t, err)

	mStruct, err := c.ModelStruct(&tests.Model{})
	require.NoError(t, err)

	defer func() {
		table, err := migrate.ModelsTable(mStruct)
		require.NoError(t, err)
		_ = internal.DropTables(ctx, p.ConnPool, table.Name, table.Schema)
	}()

	qc := query.NewCreator(c)

	newModel := func() *tests.Model {
		return &tests.Model{
			AttrString: "Something",
			Int:        3,
		}
	}
	models := []mapping.Model{newModel(), newModel()}
	// Insert models.
	err = qc.Query(mStruct, models...).Insert()
	require.NoError(t, err)

	assert.Len(t, models, 2)

	countAll, err := qc.Query(mStruct).Count()
	require.NoError(t, err)

	assert.Equal(t, int64(2), countAll)
}

// 	c, db := prepareIntegrateRepository(t)
//
// 	defer db.Close()
// 	defer deleteTestModelTable(t, db)
//
// 	s, err := query.NewC(c, &tests.Model{})
// 	require.NoError(t, err)
//
// 	count, err := s.Count()
// 	require.NoError(t, err)
//
// 	// there should be 4 created at 'prepareIntegrateRepository' instances
// 	assert.Equal(t, int64(len(testModelInstances)), count)
//
// 	for i := 0; i < 10; i++ {
// 		tm := &tests.Model{Int: i}
// 		s, err := query.NewC(c, tm)
// 		require.NoError(t, err)
//
// 		err = s.Create()
// 		require.NoError(t, err)
// 	}
//
// 	s, err = query.NewC(c, &tests.Model{})
// 	require.NoError(t, err)
//
// 	count, err = s.Count()
// 	require.NoError(t, err)
//
// 	// there should be 10 models + 4 created at 'prepareIntegrateRepository' instances
// 	assert.Equal(t, int64(10+len(testModelInstances)), count)
// }
