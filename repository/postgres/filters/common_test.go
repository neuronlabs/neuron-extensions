package filters

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

//go:generate neurogonesis models methods methods --format=goimports --type=QueryModel --single-file .

// queryModel is the model used for testing the queries filters.
type QueryModel struct {
	ID         int    `neuron:"type=primary"`
	StringAttr string `neuron:"type=attr"`
}

func getScope(t *testing.T) *query.Scope {
	t.Helper()

	m := mapping.New(
		mapping.WithNamingConvention(mapping.SnakeCase),
		mapping.WithDefaultDatabaseSchema("public"),
	)

	err := m.RegisterModels(&QueryModel{})
	require.NoError(t, err)

	mStruct, err := m.ModelStruct(&QueryModel{})
	require.NoError(t, err)

	return query.NewScope(mStruct)
}
