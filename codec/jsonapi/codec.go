package jsonapi

import (
	"encoding/json"
	"io"

	"github.com/neuronlabs/neuron/codec"
)

func init() {
	codec.RegisterCodec(MimeType, &Codec{Mime: MimeType})
}

var _ codec.Codec = &Codec{}

// Codec is jsonapi model
type Codec struct {
	Mime string
}

// MarshalErrors implements codec.Codec interface.
func (c *Codec) MarshalErrors(w io.Writer, errors codec.MultiError) error {
	p := ErrorsPayload{Errors: errors}
	return json.NewEncoder(w).Encode(&p)
}

// UnmarshalErrors implements codec.Codec interface.
func (c *Codec) UnmarshalErrors(r io.Reader) (codec.MultiError, error) {
	p := &ErrorsPayload{}
	if err := json.NewDecoder(r).Decode(p); err != nil {
		return nil, err
	}
	return p.Errors, nil
}

// MimeType implements codec.Codec interface.
func (c *Codec) MimeType() string {
	return c.Mime
}
