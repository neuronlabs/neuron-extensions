package cjson

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/mapping"
)

func TestMarshalModel(t *testing.T) {
	c := core.NewDefault()
	require.NoError(t, c.RegisterModels(&User{}, &Pet{}))

	now := time.Now()
	fatherNow := now.Add(-time.Hour * 24 * 365 * 20)
	u := &User{
		ID:           50,
		Name:         "Andrew",
		CreatedAt:    now,
		CreatedAtIso: now,
		Father: &User{
			ID:           10,
			Name:         "Werdna",
			CreatedAt:    fatherNow,
			CreatedAtIso: fatherNow,
		},
		FatherID: 10,
	}
	pet := &Pet{
		ID:      20,
		Name:    "Doggy",
		OwnerID: u.ID,
	}
	u.Pets = []*Pet{pet}

	enc, err := GetCodec(c).MarshalModel(u)
	require.NoError(t, err)

	expected := fmt.Sprintf(`{"id":50,"first_name":"Andrew","created_at":%d,"created_at_iso":"%s","father":{"id":10,"first_name":"Werdna","created_at":%d,"created_at_iso":"%s","father":null},"father_id":10,"pets":[{"id":20,"name":"Doggy","owner_id":50}]}`, now.Unix(), now.Format(codec.ISO8601TimeFormat), fatherNow.Unix(), fatherNow.Format(codec.ISO8601TimeFormat))
	assert.Equal(t, expected, string(enc))
}

func TestMarshalModels(t *testing.T) {
	c := core.NewDefault()
	require.NoError(t, c.RegisterModels(&User{}, &Pet{}))

	now := time.Now()
	pets := []mapping.Model{
		&Pet{
			ID:      10,
			Name:    "Kitty",
			Owner:   nil,
			OwnerID: 20,
		},
		&Pet{
			ID:   20,
			Name: "Parrot",
			Owner: &User{
				ID:           20,
				Name:         "ParrotOwner",
				CreatedAt:    now,
				CreatedAtIso: now,
			},
			OwnerID: 20,
		},
	}
	data, err := GetCodec(c).MarshalModels(pets)
	require.NoError(t, err)

	expected := fmt.Sprintf(`[{"id":10,"name":"Kitty","owner_id":20},{"id":20,"name":"Parrot","owner":{"id":20,"first_name":"ParrotOwner","created_at":%d,"created_at_iso":"%s","father":null},"owner_id":20}]`, now.Unix(), now.Format(codec.ISO8601TimeFormat))
	assert.Equal(t, expected, string(data))
}
