package cjsonapi

import (
	"encoding/json"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/mapping"
)

// MarshalModels implements neuronCodec.Codec interface.
func (c Codec) MarshalModels(models []mapping.Model, options ...codec.MarshalOption) ([]byte, error) {
	o := &codec.MarshalOptions{}
	for _, option := range options {
		option(o)
	}
	nodes, err := c.visitModels(models, o.Link)
	if err != nil {
		return nil, err
	}
	var data []byte
	if len(nodes) == 1 && o.SingleResult {
		data, err = json.Marshal(nodes[0])
	} else {
		data, err = json.Marshal(nodes)
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MarshalModels implements neuronCodec.Codec interface.
func (c Codec) MarshalModel(model mapping.Model, options ...codec.MarshalOption) ([]byte, error) {
	o := &codec.MarshalOptions{}
	for _, option := range options {
		option(o)
	}
	mStruct, err := c.c.ModelStruct(model)
	if err != nil {
		return nil, err
	}
	node, err := visitModelNode(mStruct, model, o.Link, mStruct.StructFields()...)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(node)
	if err != nil {
		return nil, err
	}

	return data, nil
}
