package jsonapi

import (
	"encoding/json"

	neuronCodec "github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/mapping"
)

// MarshalModels implements neuronCodec.jsonapiCodec interface.
func (j *jsonapiCodec) MarshalModels(models []mapping.Model, options neuronCodec.MarshalOptions) ([]byte, error) {
	nodes, err := j.visitModels(models, options.Link)
	if err != nil {
		return nil, err
	}
	var data []byte
	if len(nodes) == 1 && options.SingleResult {
		data, err = json.Marshal(nodes[0])
	} else {
		data, err = json.Marshal(nodes)
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}
