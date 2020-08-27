package jsonapi

import (
	"io"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

// UnmarshalPayload implements codec.PayloadUnmarshaler interface.
func (c Codec) UnmarshalPayload(r io.Reader, options ...codec.UnmarshalOption) (*codec.Payload, error) {
	o, err := c.unmarshalOptions(options)
	if err != nil {
		return nil, err
	}

	payloader, err := unmarshalPayload(r, o)
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

	payload := &codec.Payload{ModelStruct: o.ModelStruct}
	for _, node := range payloader.GetNodes() {
		if o.ModelStruct == nil {
			mStruct, ok := c.c.ModelMap.GetByCollection(node.Type)
			if !ok {
				return nil, errors.WrapDetf(codec.ErrUnmarshal, "provided unknown collection type: %s", node.Type).
					WithDetailf("provided unknown collection type: %s", node.Type)
			}
			if payload.ModelStruct == nil {
				payload.ModelStruct = mStruct
			} else if payload.ModelStruct != mStruct {
				return nil, errors.WrapDetf(codec.ErrUnmarshal, "provided payload has multiple collection data").
					WithDetail("provided payload has multiple collection data")
			}
		}
		modelStruct := payload.ModelStruct
		model := mapping.NewModel(modelStruct)
		fieldSet, err := unmarshalNode(modelStruct, node, model, includes, o)
		if err != nil {
			return nil, err
		}
		payload.Data = append(payload.Data, model)
		payload.FieldSets = append(payload.FieldSets, fieldSet)
	}
	return payload, nil
}

func (c Codec) unmarshalOptions(options []codec.UnmarshalOption) (*codec.UnmarshalOptions, error) {
	o := &codec.UnmarshalOptions{}
	for _, option := range options {
		option(o)
	}
	if o.ModelStruct == nil && o.Model == nil {
		return o, nil
	}

	if o.Model != nil {
		// Get the model struct from provided options model.
		mStruct, err := c.c.ModelStruct(o.Model)
		if err != nil {
			if errors.Is(err, mapping.ErrModelNotFound) {
				return nil, errors.Wrap(codec.ErrOptions, "provided option model is not found within given controller")
			}
			return nil, errors.Wrapf(codec.ErrOptions, "getting model struct for provided option model failed: %v", err)
		}
		if o.ModelStruct != nil && o.ModelStruct != mStruct {
			return nil, errors.Wrap(codec.ErrOptions, "provided both model and model struct options. But these values doesn't specify the same model structure")
		}
		o.ModelStruct = mStruct
	}
	return o, nil
}
