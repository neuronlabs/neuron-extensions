package collections

import (
	"testing"

	"github.com/google/uuid"
	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/database"
	"github.com/stretchr/testify/assert"
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

	uc := &NRN_Users{}
	mStruct, err := uc.modelStruct(database.New(c))
	require.NoError(t, err)

	assert.NotNil(t, uc.mStruct)
	assert.Equal(t, mStruct, c.MustModelStruct(u))
}
