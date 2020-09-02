package collections

import (
	"testing"

	"github.com/google/uuid"
	"github.com/neuronlabs/neuron/repository/mockrepo"

	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-extensions/neurogonesis/internal/tests"
	"github.com/neuronlabs/neuron-extensions/neurogonesis/internal/tests/external"
)

func TestCollection(t *testing.T) {
	name := "Name field"
	u := &tests.User{ID: uuid.New(), Name: &name}
	mp := mapping.New()
	err := mp.RegisterModels(u, &tests.Car{}, &external.Model{})
	require.NoError(t, err)

	uc := &NRN_Users{}
	db, err := database.New(database.WithModelMap(mp), database.WithDefaultRepository(&mockrepo.Repository{}))
	require.NoError(t, err)
	mStruct, err := uc.modelStruct(db)
	require.NoError(t, err)

	assert.NotNil(t, uc.mStruct)
	assert.Equal(t, mStruct, mp.MustModelStruct(u))
}
