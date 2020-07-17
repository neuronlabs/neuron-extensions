package filters

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/migrate"
)

// queryModel is the model used for testing the queries filters.
type QueryModel struct {
	ID         int    `neuron:"type=primary"`
	StringAttr string `neuron:"type=attr"`
}

//go:generate neuron-generator models methods --type=QueryModel --single-file .

func getScope(t *testing.T) *query.Scope {
	t.Helper()

	m := mapping.NewModelMap(mapping.SnakeCase)

	err := m.RegisterModels(&QueryModel{})
	require.NoError(t, err)

	mStruct, ok := m.GetModelStruct(&QueryModel{})
	require.True(t, ok)

	err = migrate.PrepareModel(mStruct)
	require.NoError(t, err)

	return query.NewScope(nil, mStruct)
}
