package jsonapi

import (
	"encoding/json"
	"io"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/controller"
)

var c *jsonapiCodec

func init() {
	c = &jsonapiCodec{Mime: MimeType}
	codec.RegisterCodec(MimeType, c)
}

// Codec gets the jsonapiCodec value.
func Codec() codec.Codec {
	return c
}

var _ codec.Codec = c

// jsonapiCodec is jsonapi model
type jsonapiCodec struct {
	Mime string
	c    *controller.Controller
}

// Initialize implements core.Initializer.
func (j *jsonapiCodec) Initialize(ctrl *controller.Controller) error {
	j.c = ctrl
	return nil
}

// MarshalErrors implements neuronCodec.jsonapiCodec interface.
func (j *jsonapiCodec) MarshalErrors(w io.Writer, errors ...*codec.Error) error {
	p := ErrorsPayload{Errors: errors}
	return json.NewEncoder(w).Encode(&p)
}

// UnmarshalErrors implements neuronCodec.jsonapiCodec interface.
func (j *jsonapiCodec) UnmarshalErrors(r io.Reader) (codec.MultiError, error) {
	p := &ErrorsPayload{}
	if err := json.NewDecoder(r).Decode(p); err != nil {
		return nil, err
	}
	return p.Errors, nil
}

// MimeType implements neuronCodec.jsonapiCodec interface.
func (j *jsonapiCodec) MimeType() string {
	return j.Mime
}
