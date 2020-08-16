package jsonapi

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
)

// TestUnmarshalScopeOne tests unmarshal scope one function.
func TestUnmarshalScopeOne(t *testing.T) {
	c := controller.NewDefault()
	err := c.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.Nil(t, err)

	cd := &Codec{c: c}
	// Case 1:
	// Correct with  attributes
	t.Run("valid_attributes", func(t *testing.T) {
		in := strings.NewReader("{\"data\": {\"type\": \"blogs\", \"id\": \"1\", \"attributes\": {\"title\": \"Some title.\"}}}")

		payload, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.NoError(t, err)

		if assert.Len(t, payload.Data, 1) {
			blog, ok := payload.Data[0].(*Blog)
			require.True(t, ok)
			assert.Equal(t, 1, blog.ID)
			assert.Equal(t, "Some title.", blog.Title)
		}
	})

	// Case 2
	// Valid with relationships and attributes
	t.Run("valid_rel_attrs", func(t *testing.T) {
		in := strings.NewReader(`
		{
			"data":{
				"type":"blogs",
				"id":"2",
				"attributes": {
					"title":"Correct Unmarshal"
				},
				"relationships":{
					"current_post":{
						"data":{
							"type":"posts",
							"id":"2"
						}					
					}
				}
			}
		}`)

		payload, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.NoError(t, err)

		if assert.Len(t, payload.Data, 1) {
			blog, ok := payload.Data[0].(*Blog)
			require.True(t, ok)
			assert.Equal(t, 2, blog.ID)
			assert.Equal(t, "Correct Unmarshal", blog.Title)

			if assert.NotNil(t, blog.CurrentPost) {
				assert.Equal(t, uint64(2), blog.CurrentPost.ID)
			}
		}
	})

	// Case 3:
	// Invalid document - no opening bracket.
	t.Run("invalid_document", func(t *testing.T) {
		in := strings.NewReader(`"data":{"type":"blogs","id":"1"}`)
		_, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, codec.ErrUnmarshal))
	})

	// Case 3 :
	// Invalid collection - unrecognized collection
	t.Run("invalid_collection", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"unrecognized","id":"1"}}`)
		_, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, codec.ErrUnmarshal))
	})

	// Case 4
	// Invalid syntax - syntax error
	t.Run("invalid_syntax", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1",}}`)
		_, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, codec.ErrUnmarshal))
	})

	// Case 5:
	// Invalid Field - unrecognized field
	t.Run("invalid_field_value", func(t *testing.T) {
		// number instead of string
		in := strings.NewReader(`{"data":{"type":"blogs","id":1.03}}`)
		_, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, codec.ErrUnmarshal))
	})

	t.Run("invalid_relationship_type", func(t *testing.T) {
		// string instead of object
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1", "relationships":"invalid"}}`)
		_, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, codec.ErrUnmarshal))
	})

	// array
	t.Run("invalid_id_value_array", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"blogs","id":{"1":"2"}}}`)
		_, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, codec.ErrUnmarshal))
	})

	// array
	t.Run("invalid_relationship_value_array", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1", "relationships":["invalid"]}}`)
		_, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, codec.ErrUnmarshal))
	})

	// bool
	t.Run("invalid_relationship_value_bool", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1", "relationships":true}}`)
		_, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, codec.ErrUnmarshal))
	})

	// Case 6:
	// invalid field value within i.e. for attribute
	t.Run("invalid_attribute_value", func(t *testing.T) {
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1", "attributes":{"title":1.02}}}`)
		_, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, codec.ErrUnmarshalFieldValue))
	})

	t.Run("invalid_field_strict_mode", func(t *testing.T) {
		// title attribute is missspelled as 'Atitle'
		in := strings.NewReader(`{"data":{"type":"blogs","id":"1", "attributes":{"Atitle":1.02}}}`)
		_, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{
			StrictUnmarshal: true,
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, codec.ErrUnmarshalFieldName))
	})

	t.Run("nil_ptr_attributes", func(t *testing.T) {
		in := strings.NewReader(`
				{
				  "data": {
				  	"type":"unmarshal_models",
				  	"id":"3",
				  	"attributes":{
				  	  "ptr_string": null,
				  	  "ptr_time": null,
				  	  "string_slice": []				  	  
				  	}
				  }
				}`)

		c := controller.NewDefault()
		err := c.RegisterModels(&UnmarshalModel{})
		require.Nil(t, err)

		cd := &Codec{c: c}

		payload, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.NoError(t, err)

		if assert.Len(t, payload.Data, 1) {
			m, ok := payload.Data[0].(*UnmarshalModel)
			if assert.True(t, ok) {
				assert.Nil(t, m.PtrString)
				assert.Nil(t, m.PtrTime)
				assert.Empty(t, m.StringSlice)
			}
		}
	})

	t.Run("ptr_attr_with_values", func(t *testing.T) {
		in := strings.NewReader(`
				{
				  "data": {
				  	"type":"unmarshal_models",
				  	"id":"3",
				  	"attributes":{
				  	  "ptr_string": "maciej",
				  	  "ptr_time": 1540909418248,
				  	  "string_slice": ["marcin","michal"]				  	  
				  	}
				  }
				}`)
		c := controller.NewDefault()
		cd := &Codec{c: c}
		err := c.RegisterModels(&UnmarshalModel{})
		require.Nil(t, err)

		payload, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		require.NoError(t, err)

		if assert.Len(t, payload.Data, 1) {
			m, ok := payload.Data[0].(*UnmarshalModel)
			if assert.True(t, ok) {
				if assert.NotNil(t, m.PtrString) {
					assert.Equal(t, "maciej", *m.PtrString)
				}
				if assert.NotNil(t, m.PtrTime) {
					assert.Equal(t, int64(1540909418248), m.PtrTime.Unix())
				}
				if assert.Len(t, m.StringSlice, 2) {
					assert.Equal(t, "marcin", m.StringSlice[0])
					assert.Equal(t, "michal", m.StringSlice[1])
				}
			}
		}
	})

	t.Run("slice_attr_with_null", func(t *testing.T) {
		in := strings.NewReader(`
				{
				  "data": {
				  	"type":"unmarshal_models",
				  	"id":"3",
				  	"attributes":{				  	  				  	  
				  	  "string_slice": [null,"michal"]				  	  
				  	}
				  }
				}`)
		c := controller.NewDefault()
		cd := &Codec{c: c}
		err := c.RegisterModels(&UnmarshalModel{})
		require.Nil(t, err)

		_, err = cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		assert.Error(t, err)
	})

	t.Run("slice_value_with_invalid_type", func(t *testing.T) {
		in := strings.NewReader(`
				{
				  "data": {
				  	"type":"unmarshal_models",
				  	"id":"3",
				  	"attributes":{				  	  				  	  
				  	  "string_slice": [1, "15"]				  	  
				  	}
				  }
				}`)
		c := controller.NewDefault()
		cd := &Codec{c: c}
		err := c.RegisterModels(&UnmarshalModel{})
		require.Nil(t, err)

		_, err = cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
		assert.Error(t, err)
	})

	t.Run("Array", func(t *testing.T) {
		t.Run("TooManyValues", func(t *testing.T) {
			c := controller.NewDefault()
			cd := &Codec{c: c}
			err := c.RegisterModels(&ArrModel{})
			require.Nil(t, err)

			in := strings.NewReader(`{"data":{"type":"arr_models","id":"1","attributes":{"arr": [1.251,125.162,16.162]}}}`)

			_, err = cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
			assert.Error(t, err)
		})

		t.Run("Correct", func(t *testing.T) {
			c := controller.NewDefault()
			cd := &Codec{c: c}
			err := c.RegisterModels(&ArrModel{})
			require.NoError(t, err)

			in := strings.NewReader(`{"data":{"type":"arr_models","id":"1","attributes":{"arr": [1.251,125.162]}}}`)

			payload, err := cd.UnmarshalPayload(in, codec.UnmarshalOptions{})
			require.NoError(t, err)

			if assert.Len(t, payload.Data, 1) {
				m, ok := payload.Data[0].(*ArrModel)
				require.True(t, ok)

				assert.Equal(t, 1, m.ID)
				if assert.Len(t, m.Arr, 2) {
					assert.Equal(t, 1.251, m.Arr[0])
					assert.Equal(t, 125.162, m.Arr[1])
				}
			}
		})
	})
	// type maptest struct {
	// 	model interface{}
	// 	r     string
	// 	f     func(t *testing.T, s *query.Scope, err error)
	// }
	// t.Run("Map", func(t *testing.T) {
	// 	t.Helper()
	//
	// type MpString struct {
	// 	ID  int               `neuron:"type=primary"`
	// 	Map map[string]string `neuron:"type=attr"`
	// }
	// type MpPtrString struct {
	// 	ID  int                `neuron:"type=primary"`
	// 	Map map[string]*string `neuron:"type=attr"`
	// }
	// type MpInt struct {
	// 	ID  int            `neuron:"type=primary"`
	// 	Map map[string]int `neuron:"type=attr"`
	// }
	// type MpPtrInt struct {
	// 	ID  int             `neuron:"type=primary"`
	// 	Map map[string]*int `neuron:"type=attr"`
	// }
	// type MpFloat struct {
	// 	ID  int                `neuron:"type=primary"`
	// 	Map map[string]float64 `neuron:"type=attr"`
	// }
	//
	// type MpPtrFloat struct {
	// 	ID  int                 `neuron:"type=primary"`
	// 	Map map[string]*float64 `neuron:"type=attr"`
	// }
	//
	// type MpSliceInt struct {
	// 	ID  int              `neuron:"type=primary"`
	// 	Map map[string][]int `neuron:"type=attr"`
	// }
	//
	// type MpSlicePtrInt struct {
	// 	ID  int               `neuron:"type=primary"`
	// 	Map map[string][]*int `neuron:"type=attr"`
	// }
	//
	// type MpSliceTime struct {
	// 	ID  int                    `neuron:"type=primary"`
	// 	Map map[string][]time.Time `neuron:"type=attr"`
	// }
	// type MpSlicePtrTime struct {
	// 	ID  int                     `neuron:"type=primary"`
	// 	Map map[string][]*time.Time `neuron:"type=attr"`
	// }
	// type MpPtrSliceTime struct {
	// 	ID  int                     `neuron:"type=primary"`
	// 	Map map[string]*[]time.Time `neuron:"type=attr"`
	// }
	// type MpArrayFloat struct {
	// 	ID  int                   `neuron:"type=primary"`
	// 	Map map[string][2]float64 `neuron:"type=attr"`
	// }
	//
	// tests := map[string]maptest{
	// 	"InvalidKey": {
	// 		model: &MpString{},
	// 		r: `{"data":{"type":"mp_strings","id":"1",
	// 	"attributes":{"map": {1:"some"}}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			assert.Error(t, err)
	// 		},
	// 	},
	// 	"StringKey": {
	// 		model: &MpString{},
	// 		r: `{"data":{"type":"mp_strings","id":"1",
	// 	"attributes":{"map": {"key":"value"}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			if assert.NoError(t, err) {
	// 				model, ok := s.Value.(*MpString)
	// 				require.True(t, ok)
	// 				if assert.NotNil(t, model.Map) {
	// 					assert.Equal(t, "value", model.Map["key"])
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"InvalidStrValue": {
	// 		model: &MpString{},
	// 		r: `{"data":{"type":"mp_strings","id":"1",
	// 	"attributes":{"map": {"key":{}}}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			assert.Error(t, err)
	// 		},
	// 	},
	// 	"InvalidStrValueFloat": {
	// 		model: &MpString{},
	// 		r: `{"data":{"type":"mp_strings","id":"1",
	// 	"attributes":{"map": {"key":1.23}}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			assert.Error(t, err)
	// 		},
	// 	},
	// 	"InvalidStrValueNil": {
	// 		model: &MpString{},
	// 		r: `{"data":{"type":"mp_strings","id":"1",
	// 	"attributes":{"map": {"key":null}}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			assert.Error(t, err)
	// 		},
	// 	},
	//
	// 	"PtrStringKey": {
	// 		model: &MpPtrString{},
	// 		r: `{"data":{"type":"mp_ptr_strings","id":"1",
	// 	"attributes":{"map": {"key":"value"}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			if assert.NoError(t, err) {
	// 				model, ok := s.Value.(*MpPtrString)
	// 				require.True(t, ok)
	// 				if assert.NotNil(t, model.Map) {
	// 					if assert.NotNil(t, model.Map["key"]) {
	// 						assert.Equal(t, "value", *model.Map["key"])
	// 					}
	//
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"NullPtrStringKey": {
	// 		model: &MpPtrString{},
	// 		r: `{"data":{"type":"mp_ptr_strings","id":"1",
	// 	"attributes":{"map": {"key":null}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			if assert.NoError(t, err) {
	// 				model, ok := s.Value.(*MpPtrString)
	// 				require.True(t, ok)
	// 				if assert.NotNil(t, model.Map) {
	// 					v, ok := model.Map["key"]
	// 					assert.True(t, ok)
	// 					assert.Nil(t, model.Map["key"], v)
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"IntKey": {
	// 		model: &MpInt{},
	// 		r: `{"data":{"type":"mp_ints","id":"1",
	// 	"attributes":{"map": {"key":1}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			if assert.NoError(t, err) {
	// 				model, ok := s.Value.(*MpInt)
	// 				require.True(t, ok)
	// 				if assert.NotNil(t, model.Map) {
	// 					v, ok := model.Map["key"]
	// 					assert.True(t, ok)
	// 					assert.Equal(t, 1, v)
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"PtrIntKey": {
	// 		model: &MpPtrInt{},
	// 		r: `{"data":{"type":"mp_ptr_ints","id":"1",
	// 	"attributes":{"map": {"key":1}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			if assert.NoError(t, err) {
	// 				model, ok := s.Value.(*MpPtrInt)
	// 				require.True(t, ok)
	// 				if assert.NotNil(t, model.Map) {
	// 					v, ok := model.Map["key"]
	// 					if assert.True(t, ok) {
	// 						if assert.NotNil(t, v) {
	// 							assert.Equal(t, 1, *v)
	// 						}
	// 					}
	//
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"NilPtrIntKey": {
	// 		model: &MpPtrInt{},
	// 		r: `{"data":{"type":"mp_ptr_ints","id":"1",
	// 	"attributes":{"map": {"key":null}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			if assert.NoError(t, err) {
	// 				model, ok := s.Value.(*MpPtrInt)
	// 				require.True(t, ok)
	// 				if assert.NotNil(t, model.Map) {
	// 					v, ok := model.Map["key"]
	// 					if assert.True(t, ok) {
	// 						assert.Nil(t, v)
	// 					}
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"FloatKey": {
	// 		model: &MpFloat{},
	// 		r: `{"data":{"type":"mp_floats","id":"1",
	// 	"attributes":{"map": {"key":1.2151}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			if assert.NoError(t, err) {
	// 				model, ok := s.Value.(*MpFloat)
	// 				require.True(t, ok)
	// 				if assert.NotNil(t, model.Map) {
	// 					v, ok := model.Map["key"]
	// 					if assert.True(t, ok) {
	// 						assert.Equal(t, 1.2151, v)
	// 					}
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"PtrFloatKey": {
	// 		model: &MpPtrFloat{},
	// 		r: `{"data":{"type":"mp_ptr_floats","id":"1",
	// 	"attributes":{"map": {"key":1.2151}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			if assert.NoError(t, err) {
	// 				model, ok := s.Value.(*MpPtrFloat)
	// 				require.True(t, ok)
	// 				if assert.NotNil(t, model.Map) {
	// 					v, ok := model.Map["key"]
	// 					if assert.True(t, ok) {
	// 						if assert.NotNil(t, v) {
	// 							assert.Equal(t, 1.2151, *v)
	// 						}
	// 					}
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"NilPtrFloatKey": {
	// 		model: &MpPtrFloat{},
	// 		r: `{"data":{"type":"mp_ptr_floats","id":"1",
	// 	"attributes":{"map": {"key":null}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			if assert.NoError(t, err) {
	// 				model, ok := s.Value.(*MpPtrFloat)
	// 				require.True(t, ok)
	// 				if assert.NotNil(t, model.Map) {
	// 					v, ok := model.Map["key"]
	// 					if assert.True(t, ok) {
	// 						assert.Nil(t, v)
	// 					}
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"InvalidMapForm": {
	// 		model: &MpPtrFloat{},
	// 		r: `{"data":{"type":"mp_ptr_floats","id":"1",
	// 		"attributes":{"map": ["string1"]}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			assert.Error(t, err)
	// 		},
	// 	},
	// 	"SliceInt": {
	// 		model: &MpSliceInt{},
	// 		r:     `{"data":{"type":"mp_slice_ints","id":"1","attributes":{"map":{"key":[1,3]}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			require.NoError(t, err)
	//
	// 			v, ok := s.Value.(*MpSliceInt)
	// 			if assert.True(t, ok) {
	// 				assert.Contains(t, v.Map["key"], 1)
	// 				assert.Contains(t, v.Map["key"], 3)
	// 			}
	// 		},
	// 	},
	//
	// 	"SlicePtrInt": {
	// 		model: &MpSlicePtrInt{},
	// 		r:     `{"data":{"type":"mp_slice_ptr_ints","id":"1","attributes":{"map":{"key":[1,3]}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			require.NoError(t, err)
	//
	// 			v, ok := s.Value.(*MpSlicePtrInt)
	// 			if assert.True(t, ok) {
	// 				kv, ok := v.Map["key"]
	// 				if assert.True(t, ok) {
	// 					var count int
	// 					for _, i := range kv {
	// 						if i != nil {
	// 							switch *i {
	// 							case 1, 3:
	// 								count++
	// 							}
	// 						}
	// 					}
	// 					assert.Equal(t, 2, count)
	// 				}
	// 				// assert.Contains(t, v.Map["key"], 1)
	// 				// assert.Contains(t, v.Map["key"], 3)
	// 			}
	// 		},
	// 	},
	// 	"SliceTime": {
	// 		model: &MpSliceTime{},
	// 		r:     `{"data":{"type":"mp_slice_times","id":"1","attributes":{"map":{"key":[1257894000,1257895000]}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			require.NoError(t, err)
	//
	// 			v, ok := s.Value.(*MpSliceTime)
	// 			if assert.True(t, ok) {
	// 				kv, ok := v.Map["key"]
	// 				if assert.True(t, ok) {
	// 					var count int
	//
	// 					for _, i := range kv {
	// 						switch i.Unix() {
	// 						case 1257895000, 1257894000:
	// 							count++
	// 						}
	// 					}
	// 					assert.Equal(t, 2, count)
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"SlicePtrTime": {
	// 		model: &MpSlicePtrTime{},
	// 		r:     `{"data":{"type":"mp_slice_ptr_times","id":"1","attributes":{"map":{"key":[1257894000,1257895000, null]}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			require.NoError(t, err)
	//
	// 			v, ok := s.Value.(*MpSlicePtrTime)
	// 			if assert.True(t, ok) {
	// 				kv, ok := v.Map["key"]
	// 				if assert.True(t, ok) {
	// 					var count int
	// 					for _, i := range kv {
	// 						if i == nil {
	// 							count++
	// 						} else {
	// 							switch i.Unix() {
	// 							case 1257895000, 1257894000:
	// 								count++
	// 							}
	// 						}
	// 					}
	// 					assert.Equal(t, 3, count)
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"PtrSliceTime": {
	// 		model: &MpPtrSliceTime{},
	// 		r:     `{"data":{"type":"mp_ptr_slice_times","id":"1","attributes":{"map":{"key":[1257894000,1257895000]}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			require.NoError(t, err)
	//
	// 			v, ok := s.Value.(*MpPtrSliceTime)
	// 			if assert.True(t, ok) {
	// 				kv, ok := v.Map["key"]
	// 				if assert.True(t, ok) && assert.NotNil(t, kv) {
	// 					var count int
	//
	// 					for _, i := range *kv {
	// 						switch i.Unix() {
	// 						case 1257895000, 1257894000:
	// 							count++
	// 						}
	// 					}
	// 					assert.Equal(t, 2, count)
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"ArrayFloat": {
	// 		model: &MpArrayFloat{},
	// 		r:     `{"data":{"type":"mp_array_floats","id":"1","attributes":{"map":{"key":[12.51,261.123]}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			require.NoError(t, err)
	//
	// 			v, ok := s.Value.(*MpArrayFloat)
	// 			if assert.True(t, ok) {
	// 				kv, ok := v.Map["key"]
	// 				if ok {
	// 					assert.InDelta(t, 12.51, kv[0], 0.01)
	// 					assert.InDelta(t, 261.123, kv[1], 0.001)
	// 				}
	// 			}
	// 		},
	// 	},
	// 	"ArrayFloatTooManyValues": {
	// 		model: &MpArrayFloat{},
	// 		r:     `{"data":{"type":"mp_array_floats","id":"1","attributes":{"map":{"key":[12.51,261.123,12.671]}}}}`,
	// 		f: func(t *testing.T, s *query.Scope, err error) {
	// 			require.Error(t, err)
	// 		},
	// 	},
	// }
	//
	// for name, test := range tests {
	// 	t.Run(name, func(t *testing.T) {
	// 		t.Helper()
	// 		c := defaultTestingController(t)
	//
	// 		in := strings.NewReader(test.r)
	// 		err := c.RegisterModels(test.model)
	// 		require.NoError(t, err)
	//
	// 		var is *query.Scope
	// 		// require.NotPanics(t, func() {
	// 		var s *query.Scope
	// 		s, err = UnmarshalSingleScopeC(c, in, test.model)
	// 		if s != nil {
	// 			is = s
	// 		}
	// 		// })
	// 		test.f(t, is, err)
	// 	})
	// }
	// })

	// t.Run("NestedStruct", func(t *testing.T) {
	// 	type NestedModel struct {
	// 		ValueFirst  int `neuron:"name=first"`
	// 		ValueSecond int `neuron:"name=second"`
	// 	}
	//
	// 	type Simple struct {
	// 		ID     int          `neuron:"type=primary"`
	// 		Nested *NestedModel `neuron:"type=attr"`
	// 	}
	//
	// 	type DoubleNested struct {
	// 		Nested *NestedModel `neuron:"name=nested"`
	// 	}
	//
	// 	type SimpleDouble struct {
	// 		ID           int           `neuron:"type=primary"`
	// 		DoubleNested *DoubleNested `neuron:"type=attr;name=double"`
	// 	}
	//
	// 	tests := map[string]maptest{
	// 		"Simple": {
	// 			r:     `{"data":{"type":"simples","attributes":{"nested":{"first":1,"second":2}}}}`,
	// 			model: &Simple{},
	// 			f: func(t *testing.T, s *query.Scope, err error) {
	// 				assert.NoError(t, err)
	// 				v, ok := s.Value.(*Simple)
	// 				if assert.True(t, ok) {
	// 					if assert.NotNil(t, v.Nested) {
	// 						assert.Equal(t, 1, v.Nested.ValueFirst)
	// 						assert.Equal(t, 2, v.Nested.ValueSecond)
	// 					}
	// 				}
	// 			},
	// 		},
	// 		"SimpleWithDoubleNested": {
	// 			r:     `{"data":{"type":"simple_doubles","attributes":{"double":{"nested":{"first":1,"second":2}}}}}`,
	// 			model: &SimpleDouble{},
	// 			f: func(t *testing.T, s *query.Scope, err error) {
	// 				assert.NoError(t, err)
	// 				v, ok := s.Value.(*SimpleDouble)
	// 				if assert.True(t, ok) {
	// 					if assert.NotNil(t, v.DoubleNested) {
	// 						nested := v.DoubleNested.Nested
	//
	// 						if assert.NotNil(t, nested) {
	// 							assert.Equal(t, 1, nested.ValueFirst)
	// 							assert.Equal(t, 2, nested.ValueSecond)
	// 						}
	// 					}
	// 				}
	// 			},
	// 		},
	// 	}
	//
	// 	for name, test := range tests {
	// 		t.Run(name, func(t *testing.T) {
	// 			c := defaultTestingController(t)
	//
	// 			in := strings.NewReader(test.r)
	// 			err := c.RegisterModels(test.model)
	// 			require.NoError(t, err)
	// 			s, err := UnmarshalSingleScopeC(c, in, test.model)
	// 			test.f(t, s, err)
	// 		})
	//
	// 	}
	// })
	//
	// t.Run("Slices", func(t *testing.T) {
	//
	// 	type AttrArrStruct struct {
	// 		ID  int       `neuron:"type=primary"`
	// 		Arr []*string `neuron:"type=attr"`
	// 	}
	//
	// 	type ArrayModel struct {
	// 		ID  int       `neuron:"type=primary"`
	// 		Arr [2]string `neuron:"type=attr"`
	// 	}
	//
	// 	type SliceInt struct {
	// 		ID int   `neuron:"type=primary"`
	// 		Sl []int `neuron:"type=attr"`
	// 	}
	//
	// 	type ArrInt struct {
	// 		ID  int    `neuron:"type=primary"`
	// 		Arr [2]int `neuron:"type=attr"`
	// 	}
	//
	// 	type NestedStruct struct {
	// 		Name string
	// 	}
	//
	// 	type SliceStruct struct {
	// 		ID int             `neuron:"type=primary"`
	// 		Sl []*NestedStruct `neuron:"type=attr"`
	// 	}
	//
	// 	tests := map[string]maptest{
	// 		"StringPtr": {
	// 			model: &AttrArrStruct{},
	// 			f: func(t *testing.T, s *query.Scope, err error) {
	// 				assert.NoError(t, err)
	// 			},
	// 			r: `{"data":{"type":"attr_arr_structs","id":"1","attributes":{"arr":["first",null,"second"]}}}`,
	// 		},
	// 		"StringArray": {
	// 			model: &ArrayModel{},
	// 			f: func(t *testing.T, s *query.Scope, err error) {
	// 				assert.NoError(t, err)
	// 			},
	// 			r: `{"data":{"type":"array_models","attributes":{"arr":["first","second"]}}}`,
	// 		},
	// 		"StringArrayOutOfRange": {
	// 			model: &ArrayModel{},
	// 			f: func(t *testing.T, s *query.Scope, err error) {
	// 				assert.Error(t, err)
	// 			},
	// 			r: `{"data":{"type":"array_models","attributes":{"arr":["first","second","third"]}}}`,
	// 		},
	// 		"IntSlice": {
	// 			model: &SliceInt{},
	// 			f: func(t *testing.T, s *query.Scope, err error) {
	// 				assert.NoError(t, err)
	// 			},
	// 			r: `{"data":{"type":"slice_ints","attributes":{"sl":[1,5]}}}`,
	// 		},
	// 		"IntSliceInvalidType": {
	// 			model: &SliceInt{},
	// 			f: func(t *testing.T, s *query.Scope, err error) {
	// 				assert.Error(t, err)
	// 			},
	// 			r: `{"data":{"type":"slice_ints","attributes":{"sl":[1,5,"string"]}}}`,
	// 		},
	// 		"StructSlice": {
	// 			model: &SliceStruct{},
	// 			f: func(t *testing.T, s *query.Scope, err error) {
	// 				assert.NoError(t, err)
	// 			},
	// 			r: `{"data":{"type":"slice_structs","attributes":{"sl":[{"name":"first"}]}}}`,
	// 		},
	// 		"IntArray": {
	// 			model: &ArrInt{},
	// 			f: func(t *testing.T, s *query.Scope, err error) {
	// 				assert.NoError(t, err)
	// 			},
	// 			r: `{"data":{"type":"arr_ints","attributes":{"arr":[1,2]}}}`,
	// 		},
	// 	}
	//
	// 	for name, test := range tests {
	// 		t.Run(name, func(t *testing.T) {
	// 			c := defaultTestingController(t)
	//
	// 			in := strings.NewReader(test.r)
	// 			var err error
	//
	// 			require.NotPanics(t, func() {
	// 				err = c.RegisterModels(test.model)
	// 				require.NoError(t, err)
	// 			})
	//
	// 			var is *query.Scope
	// 			require.NotPanics(t, func() {
	// 				var s *query.Scope
	// 				s, err = UnmarshalSingleScopeC(c, in, test.model)
	// 				if s != nil {
	// 					is = s
	// 				}
	// 			})
	//
	// 			test.f(t, is, err)
	// 		})
	// 	}
	//
	// })
}

// func TestUnmarshalScopeMany(t *testing.T) {
// 	c := defaultTestingController(t)
//
// 	require.NoError(t, c.RegisterModels(&Blog{}, &Post{}, &Comment{}))
//
// 	// Case 1:
// 	// Correct with  attributes
// 	t.Run("valid_attributes", func(t *testing.T) {
// 		in := strings.NewReader("{\"data\": [{\"type\": \"blogs\", \"id\": \"1\", \"attributes\": {\"title\": \"Some title.\"}}]}")
// 		blogs := []*Blog{}
// 		s, err := UnmarshalManyScopeC(c, in, &blogs)
// 		assert.NoError(t, err)
//
// 		if assert.NotNil(t, s) {
// 			assert.NotEmpty(t, s.Value)
// 			assert.NotEmpty(t, blogs)
// 		}
//
// 	})
//
// 	// Case 2
// 	// Walid with relationships and attributes
//
// 	t.Run("valid_rel_attrs", func(t *testing.T) {
// 		in := strings.NewReader(`{
// 		"data":[
// 			{
// 				"type":"blogs",
// 				"id":"2",
// 				"attributes": {
// 					"title":"Correct Unmarshal"
// 				},
// 				"relationships":{
// 					"current_post":{
// 						"data":{
// 							"type":"posts",
// 							"id":"2"
// 						}
// 					}
// 				}
// 			}
// 		]
// 	}`)
//
// 		s, err := UnmarshalManyScopeC(c, in, &Blog{})
// 		assert.NoError(t, err)
// 		if assert.NotNil(t, s) {
// 			assert.NotEmpty(t, s.Value)
// 		}
// 	})
//
// }
//
// // TestUnmarshalUpdateFields test unmarshal test fields.
// func TestUnmarshalUpdateFields(t *testing.T) {
// 	c := defaultTestingController(t)
//
// 	require.NoError(t, c.RegisterModels(&Blog{}, &Post{}, &Comment{}))
//
// 	buf := bytes.NewBuffer(nil)
//
// 	t.Run("attribute", func(t *testing.T) {
// 		buf.Reset()
// 		buf.WriteString(`{"data":{"type":"blogs","id":"1", 	"attributes":{"title":"New title"}}}`)
//
// 		s, err := UnmarshalSingleScopeC(c, buf, &Blog{})
// 		assert.NoError(t, err)
// 		if assert.NotNil(t, s) {
// 			attr, _ := s.Struct().Attribute("title")
// 			assert.Contains(t, s.Fieldset, attr.NeuronName())
// 			assert.Len(t, s.Fieldset, 2)
// 		}
//
// 	})
//
// 	t.Run("multiple-attributes", func(t *testing.T) {
// 		buf.Reset()
// 		buf.WriteString(`{"data":{"type":"blogs","id":"1", "attributes":{"title":"New title","view_count":16}}}`)
//
// 		s, err := UnmarshalSingleScopeC(c, buf, &Blog{})
// 		assert.NoError(t, err)
// 		if assert.NotNil(t, s) {
// 			if assert.Equal(t, "blogs", s.Struct().Collection()) {
// 				mStruct := s.Struct()
// 				title, ok := mStruct.Attribute("title")
// 				if assert.True(t, ok) {
// 					assert.Contains(t, s.Fieldset, title.NeuronName())
// 				}
// 				vCount, ok := mStruct.Attribute("view_count")
// 				if assert.True(t, ok) {
// 					assert.Contains(t, s.Fieldset, vCount.NeuronName())
// 				}
//
// 			}
// 		}
// 	})
//
// 	t.Run("relationship-to-one", func(t *testing.T) {
// 		buf.Reset()
// 		buf.WriteString(`
// {
// 	"data":	{
// 		"type":"blogs",
// 		"id":"1",
// 		"relationships":{
// 			"current_post":{
// 				"data": {
// 					"type":"posts",
// 					"id": "3"
// 				}
// 			}
// 		}
// 	}
// }`)
//
// 		s, err := UnmarshalSingleScopeC(c, buf, &Blog{})
// 		assert.NoError(t, err)
// 		if assert.NotNil(t, s) {
// 			if assert.Equal(t, "blogs", s.Struct().Collection()) {
// 				mStruct := s.Struct()
// 				assert.Len(t, s.Fieldset, 2)
//
// 				curPost, ok := mStruct.RelationField("current_post")
// 				if assert.True(t, ok) {
// 					assert.Contains(t, s.Fieldset, curPost.NeuronName())
// 				}
// 			}
// 		}
// 	})
//
// 	t.Run("relationship-to-many", func(t *testing.T) {
// 		buf.Reset()
// 		buf.WriteString(`
// {
// 	"data":	{
// 		"type":"blogs",
// 		"id":"1",
// 		"relationships":{
// 			"posts":{
// 				"data": [
// 					{
// 						"type":"posts",
// 						"id": "3"
// 					},
// 					{
// 						"type":"posts",
// 						"id": "4"
// 					}
// 				]
// 			}
// 		}
// 	}
// }`)
//
// 		s, err := UnmarshalSingleScopeC(c, buf, &Blog{})
// 		assert.NoError(t, err)
// 		if assert.NotNil(t, s) {
// 			if assert.Equal(t, "blogs", s.Struct().Collection()) {
// 				mStruct := s.Struct()
// 				assert.Len(t, s.Fieldset, 2)
// 				posts, ok := mStruct.RelationField("posts")
// 				if assert.True(t, ok) {
// 					assert.Contains(t, s.Fieldset, posts.NeuronName())
// 				}
// 			}
// 		}
// 	})
//
// 	t.Run("mixed", func(t *testing.T) {
// 		buf.Reset()
// 		buf.WriteString(`
// {
// 	"data":	{
// 		"type":"blogs",
// 		"id":"1",
// 		"attributes":{
// 			"title":"mixed"
// 		},
// 		"relationships":{
// 			"current_post":{
// 				"data": {
// 					"type":"posts",
// 					"id": "3"
// 				}
// 			},
// 			"posts":{
// 				"data": [
// 					{
// 						"type":"posts",
// 						"id": "3"
// 					}
// 				]
// 			}
// 		}
// 	}
// }`)
//
// 		s, err := UnmarshalSingleScopeC(c, buf, &Blog{})
// 		assert.NoError(t, err)
//
// 		if assert.NotNil(t, s) {
// 			if assert.Equal(t, "blogs", s.Struct().Collection()) {
// 				mStruct := s.Struct()
// 				assert.Len(t, s.Fieldset, 4)
// 				title, ok := mStruct.Attribute("title")
// 				if assert.True(t, ok) {
// 					assert.Contains(t, s.Fieldset, title.NeuronName())
// 				}
// 				fields := []string{"current_post", "posts"}
// 				for _, field := range fields {
// 					relField, ok := mStruct.RelationField(field)
// 					if assert.True(t, ok) {
// 						assert.Contains(t, s.Fieldset, relField.NeuronName())
// 					}
// 				}
//
// 			}
// 		}
// 	})
//
// 	t.Run("null-data", func(t *testing.T) {
// 		buf.Reset()
// 		buf.WriteString(`{"data": null}`)
//
// 		v := &Blog{}
// 		err := UnmarshalC(c, buf, v)
// 		require.Error(t, err)
//
// 		cl, ok := err.(errors.ClassError)
// 		require.True(t, ok)
// 		assert.Equal(t, class.EncodingUnmarshalNoData, cl.Class())
// 	})
//
// 	t.Run("empty-data", func(t *testing.T) {
// 		buf.Reset()
// 		buf.WriteString(`{"data": []}`)
//
// 		v := []*Blog{}
// 		err := UnmarshalC(c, buf, &v)
// 		require.NoError(t, err)
//
// 		assert.Len(t, v, 0)
// 	})
// }

func TestUnmarshalEmpty(t *testing.T) {
	c := controller.NewDefault()
	cd := &Codec{c: c}
	err := c.RegisterModels(&Blog{}, &Post{}, &Comment{})
	require.Nil(t, err)
	buf := &bytes.Buffer{}

	t.Run("NullData", func(t *testing.T) {
		buf.Reset()
		buf.WriteString(`{"data": null}`)

		payload, err := cd.UnmarshalPayload(buf, codec.UnmarshalOptions{})
		require.NoError(t, err)

		assert.Len(t, payload.Data, 0)
	})

	t.Run("EmptyData", func(t *testing.T) {
		buf.Reset()
		buf.WriteString(`{"data": []}`)

		payload, err := cd.UnmarshalPayload(buf, codec.UnmarshalOptions{})
		require.NoError(t, err)

		assert.Len(t, payload.Data, 0)
	})
}
