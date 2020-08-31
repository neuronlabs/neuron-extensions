package collections

import (
	"testing"

	"github.com/google/uuid"
	"github.com/magiconair/properties/assert"
	"github.com/neuronlabs/neuron/core"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-extensions/neurogonesis/internal/tests"
	"github.com/neuronlabs/neuron-extensions/neurogonesis/internal/tests/external"
)

func TestCollection(t *testing.T) {
	name := "Name field"
	u := &tests.User{ID: uuid.New(), Name: &name}
	c := core.NewDefault()
	err := c.RegisterModels(u, &tests.Car{}, &external.Model{})
	require.NoError(t, err)

	_, err = NRN_Users.FieldsValues(u, "ID", "Name")
	require.Error(t, err)

	err = NRN_Users.Initialize(c)
	require.NoError(t, err)

	values, err := NRN_Users.FieldsValues(u, "ID", "Name")
	require.NoError(t, err)

	require.Len(t, values, 2)
	assert.Equal(t, u.ID, values[0])
	assert.Equal(t, u.Name, values[1])
}
