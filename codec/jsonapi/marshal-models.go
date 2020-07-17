package jsonapi

import (
	"encoding/json"
	"io"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// MarshalModels implements codec.Codec interface.
func (c *Codec) MarshalModels(w io.Writer, modelStruct *mapping.ModelStruct, models []mapping.Model, options *codec.MarshalOptions) error {
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

// MarshalQuery implements codec.QueryMarshaler interface.
func (c *Codec) MarshalQuery(w io.Writer, s *query.Scope, options *codec.MarshalOptions) error {
	payload, err := queryPayload(s, options)
	if err != nil {
		return err
	}
	if err = marshalPayload(w, payload); err != nil {
		return err
	}
	return nil
}
