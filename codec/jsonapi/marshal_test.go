package jsonapi

import (
	"bytes"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	neuronCodec "github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// TestMarshal tests the marshal function.
func TestMarshal(t *testing.T) {
	buf := bytes.Buffer{}
	cd := &codec{}
	prepare := func(t *testing.T, models ...mapping.Model) *mapping.ModelMap {
		t.Helper()
		c := defaultTestingController(t)

		buf.Reset()
		require.NoError(t, c.RegisterModels(models...))
		return c
	}

	prepareBlogs := func(t *testing.T) *mapping.ModelMap {
		return prepare(t, &Blog{}, &Post{}, &Comment{})
	}

	tests := map[string]func(*testing.T){
		"single": func(t *testing.T) {
			c := prepareBlogs(t)
			value := &Blog{ID: 5, Title: "My title", ViewCount: 14}
			if assert.NoError(t, cd.MarshalModels(&buf, c.MustModelStruct(value), []mapping.Model{value}, nil)) {
				marshaled := buf.String()
				assert.Contains(t, marshaled, `"title":"My title"`)
				assert.Contains(t, marshaled, `"view_count":14`)
				assert.Contains(t, marshaled, `"id":"5"`)
			}
		},
		"Time": func(t *testing.T) {

			t.Run("NoPtr", func(t *testing.T) {
				c := prepare(t, &ModelTime{})
				now := time.Now()
				value := &ModelTime{ID: 5, Time: now}

				if assert.NoError(t, cd.MarshalModels(&buf, c.MustModelStruct(value), []mapping.Model{value}, nil)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, "time")
					assert.Contains(t, marshaled, `"id":"5"`)
				}
			})

			t.Run("Ptr", func(t *testing.T) {
				c := prepare(t, &ModelPtrTime{})
				now := time.Now()
				value := &ModelPtrTime{ID: 5, Time: &now}
				if assert.NoError(t, cd.MarshalModels(&buf, c.MustModelStruct(value), []mapping.Model{value}, nil)) {
					marshaled := buf.String()
					assert.Contains(t, marshaled, "time")
					assert.Contains(t, marshaled, `"id":"5"`)
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
			c := prepareBlogs(t)

			values := []mapping.Model{&Blog{ID: 5, Title: "First"}, &Blog{ID: 2, Title: "Second"}}
			if assert.NoError(t, cd.MarshalModels(&buf, c.MustModelStruct(&Blog{}), values, nil)) {
				marshaled := buf.String()
				assert.Contains(t, marshaled, `"title":"First"`)
				assert.Contains(t, marshaled, `"title":"Second"`)

				assert.Contains(t, marshaled, `"id":"5"`)
				assert.Contains(t, marshaled, `"id":"2"`)
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

// TestMarshalScope tests marshaling the query.
func TestMarshalScope(t *testing.T) {
	t.Run("MarshalToManyRelationship", func(t *testing.T) {
		c := defaultTestingController(t)
		require.NoError(t, c.RegisterModels(&Pet{}, &User{}, &UserPets{}))

		pet := &Pet{ID: 5, Owners: []*User{{ID: 2, privateField: 1}, {ID: 3}}}
		modelStruct := c.MustModelStruct(pet)
		s := query.NewScope(modelStruct, pet)

		owners, ok := modelStruct.RelationByName("Owners")
		require.True(t, ok)

		err := s.Include(owners)
		require.NoError(t, err)

		payload, err := queryPayload(s, nil)
		if assert.NoError(t, err) {
			single, ok := payload.(*ManyPayload)
			require.True(t, ok)
			if assert.Len(t, single.Data, 1) {
				assert.Equal(t, strconv.Itoa(pet.ID), single.Data[0].ID)
				if assert.NotEmpty(t, single.Data[0].Relationships) {
					if assert.NotNil(t, single.Data[0].Relationships["owners"]) {
						owners, ok := single.Data[0].Relationships["owners"].(*RelationshipManyNode)
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
		}

	})

	t.Run("MarshalToManyEmptyRelationship", func(t *testing.T) {
		c := defaultTestingController(t)
		require.NoError(t, c.RegisterModels(&Pet{}, &User{}, &UserPets{}))

		pet := &Pet{ID: 5, Owners: []*User{}}

		modelStruct := c.MustModelStruct(pet)
		s := query.NewScope(modelStruct, pet)

		owners, ok := modelStruct.RelationByName("Owners")
		require.True(t, ok)

		err := s.Include(owners)
		require.NoError(t, err)

		payload, err := queryPayload(s, &neuronCodec.MarshalOptions{SingleResult: true})
		if assert.NoError(t, err) {
			single, ok := payload.(*SinglePayload)
			if assert.True(t, ok) {
				if assert.NotNil(t, single.Data) {
					assert.Equal(t, strconv.Itoa(pet.ID), single.Data.ID)
					if assert.NotEmpty(t, single.Data.Relationships) {
						if assert.NotNil(t, single.Data.Relationships["owners"]) {
							owners, ok := single.Data.Relationships["owners"].(*RelationshipManyNode)
							if assert.True(t, ok, reflect.TypeOf(single.Data.Relationships["owners"]).String()) {
								if assert.NotNil(t, owners) {
									assert.Empty(t, owners.Data)
								}
							}
						}
					}
				}
				buf := bytes.Buffer{}
				assert.NoError(t, marshalPayload(&buf, single))
				assert.Contains(t, buf.String(), "owners")
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

var ctrlConfig *config.Controller

func defaultTestingController(t *testing.T) *mapping.ModelMap {
	t.Helper()
	return mapping.NewModelMap(mapping.SnakeCase)
}

type CustomTagModel struct {
	ID                int
	VisibleCustomName string `codec:"visible"`
	HiddenField       bool   `codec:"-"`
	OmitEmptyField    string `codec:"omitempty"`
	CustomOmitEmpty   string `codec:"custom,omitempty"`
}

func TestMarshalCustomTag(t *testing.T) {
	c := defaultTestingController(t)
	assert.NoError(t, c.RegisterModels(&CustomTagModel{}))

	hidden := &CustomTagModel{ID: 1, VisibleCustomName: "found", HiddenField: true, OmitEmptyField: "found", CustomOmitEmpty: "found"}

	modelStruct := c.MustModelStruct(hidden)
	s := query.NewScope(modelStruct, hidden)

	attributes := modelStruct.Attributes()
	s.FieldSet = append([]*mapping.StructField{modelStruct.Primary()}, attributes...)
	payload, err := queryPayload(s, &neuronCodec.MarshalOptions{SingleResult: true})
	assert.NoError(t, err)

	buffer := &bytes.Buffer{}
	err = marshalPayload(buffer, payload)
	assert.NoError(t, err)

	sp, ok := payload.(*SinglePayload)
	require.True(t, ok)

	attrs := sp.Data.Attributes

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
