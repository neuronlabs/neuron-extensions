package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/tests"
	"github.com/neuronlabs/neuron/query"
)

// TestParseDeleteQuery tests the parse delete query method.
func TestParseDeleteQuery(t *testing.T) {
	c := testingController(t, false, &tests.Model{})
	p := testingRepository(c)

	mStruct, err := c.ModelStruct(&tests.Model{})
	require.NoError(t, err)

	s := query.NewScope(mStruct)
	s.Filters = query.Filters{
		query.NewFilterField(mStruct.Primary(), query.OpIn, 3, 10),
	}
	q, err := p.parseDeleteQuery(s)
	require.NoError(t, err)

	assert.Equal(t, "DELETE FROM public.models WHERE id IN ($1,$2)", q.query)
	assert.ElementsMatch(t, q.values, []interface{}{3, 10})
}
