package jsonapi

import (
	"encoding/json"
	"io"

	neuronCodec "github.com/neuronlabs/neuron/codec"
)

var c *codec

func init() {
	c = &codec{Mime: MimeType}
	neuronCodec.RegisterCodec(MimeType, c)
}

// Codec gets the codec value.
func Codec() neuronCodec.Codec {
	return c
}

var _ neuronCodec.Codec = c

// codec is jsonapi model
type codec struct {
	Mime string
}

// MarshalErrors implements neuronCodec.codec interface.
func (c *codec) MarshalErrors(w io.Writer, errors ...*neuronCodec.Error) error {
	p := ErrorsPayload{Errors: errors}
	return json.NewEncoder(w).Encode(&p)
}

// UnmarshalErrors implements neuronCodec.codec interface.
func (c *codec) UnmarshalErrors(r io.Reader) (neuronCodec.MultiError, error) {
	p := &ErrorsPayload{}
	if err := json.NewDecoder(r).Decode(p); err != nil {
		return nil, err
	}
	return p.Errors, nil
}

// MimeType implements neuronCodec.codec interface.
func (c *codec) MimeType() string {
	return c.Mime
}
