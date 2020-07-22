package jsonapi

import (
	"encoding/json"
	"io"

	neuronCodec "github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// MarshalModels implements neuronCodec.codec interface.
func (c *codec) MarshalModels(w io.Writer, modelStruct *mapping.ModelStruct, models []mapping.Model, options *neuronCodec.MarshalOptions) error {
	nodes, err := visitModels(modelStruct, models, options)
	if err != nil {
		log.Debug2f("visitModels failed: %v", err)
		return err
	}

	var payload Payloader
	if len(nodes) == 1 && options != nil && options.SingleResult {
		payload = &SinglePayload{Data: nodes[0]}
	} else {
		payload = &ManyPayload{Data: nodes}
	}
	if err = json.NewEncoder(w).Encode(payload); err != nil {
		return err
	}
	return nil
}

// MarshalQuery implements neuronCodec.QueryMarshaler interface.
func (c *codec) MarshalQuery(w io.Writer, s *query.Scope, options *neuronCodec.MarshalOptions) error {
	payload, err := queryPayload(s, options)
	if err != nil {
		return err
	}
	if err = marshalPayload(w, payload); err != nil {
		return err
	}
	return nil
}
