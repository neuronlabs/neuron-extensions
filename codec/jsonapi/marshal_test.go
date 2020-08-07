package jsonapi

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// TestMarshal tests the marshal function.
func TestMarshal(t *testing.T) {
	prepare := func(t *testing.T, models ...mapping.Model) *Codec {
		t.Helper()
		c := controller.NewDefault()
		require.NoError(t, c.RegisterModels(models...))
		return &Codec{c: c}
	}

	prepareBlogs := func(t *testing.T) *Codec {
		return prepare(t, &Blog{}, &Post{}, &Comment{})
	}

	tests := map[string]func(*testing.T){
		"single": func(t *testing.T) {
			cd := prepareBlogs(t)
			value := &Blog{ID: 5, Title: "My title", ViewCount: 14}
			data, err := cd.MarshalModels([]mapping.Model{value}, codec.MarshalOptions{})
			if assert.NoError(t, err) {
				assert.Contains(t, string(data), `"title":"My title"`)
				assert.Contains(t, string(data), `"view_count":14`)
				assert.Contains(t, string(data), `"id":"5"`)
			}
		},
		"Time": func(t *testing.T) {

			t.Run("NoPtr", func(t *testing.T) {
				cd := prepare(t, &ModelTime{})
				now := time.Now()
				value := &ModelTime{ID: 5, Time: now}

				marshaled, err := cd.MarshalModels([]mapping.Model{value}, codec.MarshalOptions{})
				if assert.NoError(t, err) {
					assert.Contains(t, string(marshaled), "time")
					assert.Contains(t, string(marshaled), `"id":"5"`)
				}
			})

			t.Run("Ptr", func(t *testing.T) {
				cd := prepare(t, &ModelPtrTime{})
				now := time.Now()
				value := &ModelPtrTime{ID: 5, Time: &now}
				marshaled, err := cd.MarshalModels([]mapping.Model{value}, codec.MarshalOptions{})
				if assert.NoError(t, err) {
					assert.Contains(t, string(marshaled), "time")
					assert.Contains(t, string(marshaled), `"id":"5"`)
				}
			})
		},
		// "singleWithMap": func(t *testing.T) {
		// 	t.Run("PtrString", func(t *testing.T) {
		// 		type MpString struct {
		// 			ID  int                `neuron:"type=primary"`
		// 			Map map[string]*string `neuron:"type=attr"`
		// 		}
		// 		c := prepare(t, &MpString{})
		//
		// 		kv := "some"
		// 		value := &MpString{ID: 5, Map: map[string]*string{"key": &kv}}
		// 		if assert.NoError(t, MarshalC(c, &buf, value)) {
		// 			marshaled := buf.String()
		// 			assert.Contains(t, marshaled, `"map":{"key":"some"}`)
		// 		}
		// 	})
		//
		// 	t.Run("NilString", func(t *testing.T) {
		// 		type MpString struct {
		// 			ID  int                `neuron:"type=primary"`
		// 			Map map[string]*string `neuron:"type=attr"`
		// 		}
		// 		c := prepare(t, &MpString{})
		// 		value := &MpString{ID: 5, Map: map[string]*string{"key": nil}}
		// 		if assert.NoError(t, MarshalC(c, &buf, value)) {
		// 			marshaled := buf.String()
		// 			assert.Contains(t, marshaled, `"map":{"key":null}`)
		// 		}
		// 	})
		//
		// 	t.Run("PtrInt", func(t *testing.T) {
		// 		type MpInt struct {
		// 			ID  int             `neuron:"type=primary"`
		// 			Map map[string]*int `neuron:"type=attr"`
		// 		}
		// 		c := prepare(t, &MpInt{})
		//
		// 		kv := 5
		// 		value := &MpInt{ID: 5, Map: map[string]*int{"key": &kv}}
		// 		if assert.NoError(t, MarshalC(c, &buf, value)) {
		// 			marshaled := buf.String()
		// 			assert.Contains(t, marshaled, `"map":{"key":5}`)
		// 		}
		// 	})
		// 	t.Run("NilPtrInt", func(t *testing.T) {
		// 		type MpInt struct {
		// 			ID  int             `neuron:"type=primary"`
		// 			Map map[string]*int `neuron:"type=attr"`
		// 		}
		// 		c := prepare(t, &MpInt{})
		//
		// 		value := &MpInt{ID: 5, Map: map[string]*int{"key": nil}}
		// 		if assert.NoError(t, MarshalC(c, &buf, value)) {
		// 			marshaled := buf.String()
		// 			assert.Contains(t, marshaled, `"map":{"key":null}`)
		// 		}
		// 	})
		// 	t.Run("PtrFloat", func(t *testing.T) {
		// 		type MpFloat struct {
		// 			ID  int                 `neuron:"type=primary"`
		// 			Map map[string]*float64 `neuron:"type=attr"`
		// 		}
		// 		c := prepare(t, &MpFloat{})
		//
		// 		fv := 1.214
		// 		value := &MpFloat{ID: 5, Map: map[string]*float64{"key": &fv}}
		// 		if assert.NoError(t, MarshalC(c, &buf, value)) {
		// 			marshaled := buf.String()
		// 			assert.Contains(t, marshaled, `"map":{"key":1.214}`)
		// 		}
		// 	})
		// 	t.Run("NilPtrFloat", func(t *testing.T) {
		// 		type MpFloat struct {
		// 			ID  int                 `neuron:"type=primary"`
		// 			Map map[string]*float64 `neuron:"type=attr"`
		// 		}
		// 		c := prepare(t, &MpFloat{})
		//
		// 		value := &MpFloat{ID: 5, Map: map[string]*float64{"key": nil}}
		// 		if assert.NoError(t, MarshalC(c, &buf, value)) {
		// 			marshaled := buf.String()
		// 			assert.Contains(t, marshaled, `"map":{"key":null}`)
		// 		}
		// 	})
		//
		// 	t.Run("SliceInt", func(t *testing.T) {
		// 		type MpSliceInt struct {
		// 			ID  int              `neuron:"type=primary"`
		// 			Map map[string][]int `neuron:"type=attr"`
		// 		}
		// 		c := prepare(t, &MpSliceInt{})
		//
		// 		value := &MpSliceInt{ID: 5, Map: map[string][]int{"key": {1, 5}}}
		// 		if assert.NoError(t, MarshalC(c, &buf, value)) {
		// 			marshaled := buf.String()
		// 			assert.Contains(t, marshaled, `"map":{"key":[1,5]}`)
		// 		}
		// 	})
		// },
		"many": func(t *testing.T) {
			cd := prepareBlogs(t)

			values := []mapping.Model{&Blog{ID: 5, Title: "First"}, &Blog{ID: 2, Title: "Second"}}
			marshaled, err := cd.MarshalModels(values, codec.MarshalOptions{})
			if assert.NoError(t, err) {
				assert.Contains(t, string(marshaled), `"title":"First"`)
				assert.Contains(t, string(marshaled), `"title":"Second"`)
				assert.Contains(t, string(marshaled), `"id":"5"`)
				assert.Contains(t, string(marshaled), `"id":"2"`)
			}
		},
		// "Nested": func(t *testing.T) {
		// 	t.Run("Simple", func(t *testing.T) {
		// 		type NestedSub struct {
		// 			First int
		// 		}
		//
		// 		type Simple struct {
		// 			ID     int        `neuron:"type=primary"`
		// 			Nested *NestedSub `neuron:"type=attr"`
		// 		}
		//
		// 		c := prepare(t, &Simple{})
		//
		// 		err := MarshalC(c, &buf, &Simple{ID: 2, Nested: &NestedSub{First: 1}})
		// 		if assert.NoError(t, err) {
		// 			marshaled := buf.String()
		// 			assert.Contains(t, marshaled, `"nested":{"first":1}`)
		// 		}
		// 	})
		//
		// 	t.Run("DoubleNested", func(t *testing.T) {
		//
		// 		type NestedSub struct {
		// 			First int
		// 		}
		//
		// 		type DoubleNested struct {
		// 			Nested *NestedSub
		// 		}
		//
		// 		type Simple struct {
		// 			ID     int           `neuron:"type=primary"`
		// 			Double *DoubleNested `neuron:"type=attr"`
		// 		}
		//
		// 		c := prepare(t, &Simple{})
		//
		// 		err := MarshalC(c, &buf, &Simple{ID: 2, Double: &DoubleNested{Nested: &NestedSub{First: 1}}})
		// 		if assert.NoError(t, err) {
		// 			marshaled := buf.String()
		// 			assert.Contains(t, marshaled, `"nested":{"first":1}`)
		// 			assert.Contains(t, marshaled, `"double":{"nested"`)
		// 		}
		// 	})
		//
		// },
	}

	for name, testFunc := range tests {
		t.Run(name, testFunc)
	}

}

// TestMarshalPayload tests marshaling the query.
func TestMarshalPayload(t *testing.T) {
	t.Run("MarshalToManyRelationship", func(t *testing.T) {
		c := controller.NewDefault()
		require.NoError(t, c.RegisterModels(&Pet{}, &User{}, &UserPets{}))

		pet := &Pet{ID: 5, Owners: []*User{{ID: 2, privateField: 1}, {ID: 3}}}
		modelStruct := c.MustModelStruct(pet)

		cd := &Codec{c: c}

		owners, ok := modelStruct.RelationByName("Owners")
		require.True(t, ok)

		payload := &codec.Payload{
			ModelStruct: modelStruct,
			Data:        []mapping.Model{pet},
			IncludedRelations: []*query.IncludedRelation{{
				StructField: owners,
				Fieldset:    owners.Relationship().RelatedModelStruct().StructFields(),
			}},
		}

		nodes, err := cd.visitPayloadModels(payload)
		if assert.NoError(t, err) {
			assert.Equal(t, strconv.Itoa(pet.ID), nodes[0].ID)
			if assert.NotEmpty(t, nodes[0].Relationships) {
				if assert.NotNil(t, nodes[0].Relationships["owners"]) {
					owners, ok := nodes[0].Relationships["owners"].(*RelationshipManyNode)
					if assert.True(t, ok) {
						var count int
						for _, owner := range owners.Data {
							if assert.NotNil(t, owner) {
								switch owner.ID {
								case "2", "3":
									count++
								}
							}
						}
						assert.Equal(t, 2, count)
					}
				}
			}
		}
	})

	t.Run("MarshalToManyEmptyRelationship", func(t *testing.T) {
		c := controller.NewDefault()
		require.NoError(t, c.RegisterModels(&Pet{}, &User{}, &UserPets{}))

		pet := &Pet{ID: 5, Owners: []*User{}}

		modelStruct := c.MustModelStruct(pet)

		owners, ok := modelStruct.RelationByName("Owners")
		require.True(t, ok)

		payload := &codec.Payload{
			ModelStruct: modelStruct,
			Data:        []mapping.Model{pet},
			IncludedRelations: []*query.IncludedRelation{{
				StructField: owners,
				Fieldset:    owners.Relationship().RelatedModelStruct().StructFields(),
			}},
		}
		cd := &Codec{c: c}

		nodes, err := cd.visitPayloadModels(payload)
		require.NoError(t, err)

		assert.Equal(t, strconv.Itoa(pet.ID), nodes[0].ID)
		if assert.NotEmpty(t, nodes[0].Relationships) {
			if assert.NotNil(t, nodes[0].Relationships["owners"]) {
				owners, ok := nodes[0].Relationships["owners"].(*RelationshipManyNode)
				if assert.True(t, ok, reflect.TypeOf(nodes[0].Relationships["owners"]).String()) {
					if assert.NotNil(t, owners) {
						assert.Empty(t, owners.Data)
					}
				}
			}
		}
	})
}

// func TestMarshalScopeRelationship(t *testing.T) {
// 	c := blogController(t)

// 	scope := blogScope(t, c)

// 	req := httptest.NewRequest("GET", "/blogs/1/relationships/posts", nil)
// 	scope, errs, err := c.BuildScopeRelationship(req, &Endpoint{Type: GetRelationship}, &ModelHandler{ModelType: reflect.TypeOf(Blog{})})

// 	assert.Nil(t, err)
// 	assert.Empty(t, errs)

// 	query.Value = &Blog{ID: 1, Posts: []*Post{{ID: 1}, {ID: 3}}}

// 	postsScope, err := query.GetRelationshipScope()
// 	assert.Nil(t, err)

// 	payload, err := c.MarshalScope(postsScope)
// 	assert.Nil(t, err)

// 	buffer := bytes.NewBufferString("")

// 	err = MarshalPayload(buffer, payload)
// 	assert.Nil(t, err)

// }

// // HiddenModel is the neuron model with hidden fields.
// type HiddenModel struct {
// 	ID          int    `neuron:"type=primary;flags=hidden"`
// 	Visibile    string `neuron:"type=attr"`
// 	HiddenField string `neuron:"type=attr;flags=hidden"`
// }
//
// // Collection implements CollectionNamer interface.
// func (h *HiddenModel) Collection() string {
// 	return "hiddens"
// }
//
// // TestMarshalHiddenScope tests if the marshaling scope would hide the 'hidden' field.
// func TestMarshalHiddenScope(t *testing.T) {
// 	c := defaultTestingController(t)
// 	assert.NoError(t, c.RegisterModels(&HiddenModel{}))
//
// 	hidden := &HiddenModel{ID: 1, Visibile: "Visible", HiddenField: "Invisible"}
//
// 	s, err := query.NewC(c, hidden)
// 	assert.NoError(t, err)
//
// 	payload, err := queryPayload(s)
// 	assert.NoError(t, err)
//
// 	buffer := &bytes.Buffer{}
// 	err = marshalPayload(buffer, payload)
// 	assert.NoError(t, err)
//
// }

func TestMarshalCustomTag(t *testing.T) {
	c := controller.NewDefault()
	assert.NoError(t, c.RegisterModels(&CustomTagModel{}))

	hidden := &CustomTagModel{ID: 1, VisibleCustomName: "found", HiddenField: true, OmitEmptyField: "found", CustomOmitEmpty: "found"}

	modelStruct := c.MustModelStruct(hidden)

	attributes := modelStruct.Attributes()
	payload := &codec.Payload{
		ModelStruct: modelStruct,
		Data:        []mapping.Model{hidden},
		FieldSets:   []mapping.FieldSet{append([]*mapping.StructField{modelStruct.Primary()}, attributes...)},
	}

	cd := &Codec{c: c}

	nodes, err := cd.visitPayloadModels(payload)
	require.NoError(t, err)

	require.Len(t, nodes, 1)

	attrs := nodes[0].Attributes
	atr, ok := attrs["visible"]
	if assert.True(t, ok, attrs) {
		assert.Equal(t, "found", atr)
	}

	atr, ok = attrs["omit_empty_field"]
	if assert.True(t, ok) {
		assert.Equal(t, "found", atr)
	}

	atr, ok = attrs["custom"]
	if assert.True(t, ok) {
		assert.Equal(t, "found", atr)
	}

	_, ok = attrs["hidden_field"]
	assert.False(t, ok)
}
