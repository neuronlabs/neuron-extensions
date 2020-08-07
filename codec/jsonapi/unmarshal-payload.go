package jsonapi

import (
	"io"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

// UnmarshalPayload implements codec.PayloadUnmarshaler interface.
func (c Codec) UnmarshalPayload(r io.Reader, options codec.UnmarshalOptions) (*codec.Payload, error) {
	payloader, err := unmarshalPayload(r, options)
	if err != nil {
		return nil, unmarshalHandleDecodeError(err)
	}

	var includes map[string]*Node
	if len(payloader.GetIncluded()) != 0 {
		includes = map[string]*Node{}
		for _, included := range payloader.GetIncluded() {
			includes[includedKeyFunc(included)] = included
		}
	}

	payload := &codec.Payload{ModelStruct: options.ModelStruct}
	for _, node := range payloader.GetNodes() {
		if options.ModelStruct == nil {
			mStruct, ok := c.c.ModelMap.GetByCollection(node.Type)
			if !ok {
				return nil, errors.NewDetf(codec.ClassUnmarshal, "provided unknown collection type: %s", node.Type)
			}
			if payload.ModelStruct == nil {
				payload.ModelStruct = mStruct
			} else if payload.ModelStruct != mStruct {
				return nil, errors.NewDetf(codec.ClassUnmarshal, "provided payload has multiple collection data")
			}
		}
		modelStruct := payload.ModelStruct
		model := mapping.NewModel(modelStruct)
		fieldSet, err := unmarshalNode(modelStruct, node, model, includes, options)
		if err != nil {
			return nil, err
		}
		payload.Data = append(payload.Data, model)
		payload.FieldSets = append(payload.FieldSets, fieldSet)
	}
	return payload, nil
}
