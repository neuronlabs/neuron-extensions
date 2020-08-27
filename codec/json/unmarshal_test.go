package json

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/core"
)

func TestUnmarshalModels(t *testing.T) {
	c := core.NewDefault()
	require.NoError(t, c.RegisterModels(&User{}, &Pet{}))
	now := time.Now().Truncate(time.Second).UTC()

	t.Run("Multi", func(t *testing.T) {
		data := []byte(fmt.Sprintf(`[{"id":10,"name":"Kitty","owner_id":20},{"id":20,"name":"Parrot","owner":{"id":20,"first_name":"ParrotOwner","created_at":%d,"created_at_iso":"%s","father":null},"owner_id":20}]`, now.Unix(), now.Format(codec.ISO8601TimeFormat)))
		// Unmarshal data.
		models, err := GetCodec(c).UnmarshalModels(data, codec.UnmarshalWithModel(&Pet{}))
		require.NoError(t, err)

		require.Len(t, models, 2)

		// First model should be
		//
		// &Pet{
		// 	ID:      10,
		// 	Name:    "Kitty",
		// 	Owner:   nil,
		// 	OwnerID: 20,
		// },
		p1 := models[0].(*Pet)
		assert.Equal(t, 10, p1.ID)
		assert.Equal(t, "Kitty", p1.Name)
		assert.Equal(t, 20, p1.OwnerID)
		assert.Nil(t, p1.Owner)
		// Second model should be.
		// &Pet{
		// 	ID:   20,
		// 	Name: "Parrot",
		// 	Owner: &User{
		// 		ID:           20,
		// 		Name:         "ParrotOwner",
		// 		CreatedAt:    now,
		// 		CreatedAtIso: now,
		// 	},
		// 	OwnerID: 20,
		// },
		p2 := models[1].(*Pet)
		assert.Equal(t, 20, p2.ID)
		assert.Equal(t, "Parrot", p2.Name)
		assert.Equal(t, 20, p2.OwnerID)
		if assert.NotNil(t, p2.Owner) {
			assert.Equal(t, 20, p2.Owner.ID)
			assert.Equal(t, "ParrotOwner", p2.Owner.Name)
			assert.Equal(t, now, p2.Owner.CreatedAt.UTC())
			assert.Equal(t, now, p2.Owner.CreatedAtIso.UTC())
		}
	})

	t.Run("Single", func(t *testing.T) {
		data := []byte(fmt.Sprintf(`{"id":20,"name":"Parrot","owner":{"id":20,"first_name":"ParrotOwner","created_at":%d,"created_at_iso":"%s","father":null},"owner_id":20}`, now.Unix(), now.Format(codec.ISO8601TimeFormat)))
		// Unmarshal data.
		model, err := GetCodec(c).UnmarshalModel(data, codec.UnmarshalWithModel(&Pet{}))
		require.NoError(t, err)

		// Second model should be.
		// &Pet{
		// 	ID:   20,
		// 	Name: "Parrot",
		// 	Owner: &User{
		// 		ID:           20,
		// 		Name:         "ParrotOwner",
		// 		CreatedAt:    now,
		// 		CreatedAtIso: now,
		// 	},
		// 	OwnerID: 20,
		// },
		p2 := model.(*Pet)
		assert.Equal(t, 20, p2.ID)
		assert.Equal(t, "Parrot", p2.Name)
		assert.Equal(t, 20, p2.OwnerID)
		if assert.NotNil(t, p2.Owner) {
			assert.Equal(t, 20, p2.Owner.ID)
			assert.Equal(t, "ParrotOwner", p2.Owner.Name)
			assert.Equal(t, now, p2.Owner.CreatedAt.UTC())
			assert.Equal(t, now, p2.Owner.CreatedAtIso.UTC())
		}
	})
}

func TestUnmarshalModel(t *testing.T) {
	c := core.NewDefault()
	require.NoError(t, c.RegisterModels(&User{}, &Pet{}))
	now := time.Now().Truncate(time.Second).UTC()

	data := []byte(fmt.Sprintf(`{"id":20,"name":"Parrot","owner":{"id":20,"first_name":"ParrotOwner","created_at":%d,"created_at_iso":"%s","father":null},"owner_id":20}`, now.Unix(), now.Format(codec.ISO8601TimeFormat)))
	// Unmarshal data.
	model, err := GetCodec(c).UnmarshalModel(data, codec.UnmarshalWithModel(&Pet{}))
	require.NoError(t, err)

	// Second model should be.
	// &Pet{
	// 	ID:   20,
	// 	Name: "Parrot",
	// 	Owner: &User{
	// 		ID:           20,
	// 		Name:         "ParrotOwner",
	// 		CreatedAt:    now,
	// 		CreatedAtIso: now,
	// 	},
	// 	OwnerID: 20,
	// },
	p2 := model.(*Pet)
	assert.Equal(t, 20, p2.ID)
	assert.Equal(t, "Parrot", p2.Name)
	assert.Equal(t, 20, p2.OwnerID)
	if assert.NotNil(t, p2.Owner) {
		assert.Equal(t, 20, p2.Owner.ID)
		assert.Equal(t, "ParrotOwner", p2.Owner.Name)
		assert.Equal(t, now, p2.Owner.CreatedAt.UTC())
		assert.Equal(t, now, p2.Owner.CreatedAtIso.UTC())
	}
}
