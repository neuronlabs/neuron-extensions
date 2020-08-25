package jsonapi

import (
	"encoding/json"
	"io"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/core"
)

// Codec gets the Codec value.
func GetCodec(c *core.Controller) codec.Codec {
	return Codec{c: c}
}

var _ codec.Codec = &Codec{}

// Codec is jsonapi model
type Codec struct {
	c *core.Controller
}

// MarshalErrors implements neuronCodec.Codec interface.
func (c Codec) MarshalErrors(w io.Writer, errors ...*codec.Error) error {
	p := ErrorsPayload{Errors: errors}
	return json.NewEncoder(w).Encode(&p)
}

// UnmarshalErrors implements neuronCodec.Codec interface.
func (c Codec) UnmarshalErrors(r io.Reader) (codec.MultiError, error) {
	p := &ErrorsPayload{}
	if err := json.NewDecoder(r).Decode(p); err != nil {
		return nil, err
	}
	return p.Errors, nil
}

// MimeType implements neuronCodec.Codec interface.
func (c Codec) MimeType() string {
	return MimeType
}
