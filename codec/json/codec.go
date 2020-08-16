package json

import (
	"encoding/json"
	"io"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
)

// MimeType defines the default mime type for the json codec.
const MimeType = "application/json"

// Compile time check if Codec implements codec.Codec.
var _ codec.Codec = &Codec{}

// Codec is a json codec.Codec implementation.
type Codec struct {
	c *controller.Controller
}

// GetCodec gets the codec with provided controller 'c'.
func GetCodec(c *controller.Controller) Codec {
	return Codec{c: c}
}

// MarshalErrors implements codec.Codec interface.
func (c Codec) MarshalErrors(w io.Writer, errs ...*codec.Error) error {
	data, err := json.Marshal(errs)
	if err != nil {
		return errors.Wrapf(codec.ErrMarshal, "marshaling errors failed: %v", err)
	}
	_, err = w.Write(data)
	if err != nil {
		return errors.Wrapf(codec.ErrMarshal, "writing marshaled errors failed: %v", err)
	}
	return nil
}

// UnmarshalErrors implements codec.Codec interface.
func (c Codec) UnmarshalErrors(r io.Reader) (codec.MultiError, error) {
	errs := []*codec.Error{}
	if err := json.NewDecoder(r).Decode(&errs); err != nil {
		return nil, errors.WrapDetf(codec.ErrUnmarshal, "unmarshal errors failed: %v", err)
	}
	return errs, nil
}

// MimeType implements codec.Codec interface.
func (c Codec) MimeType() string {
	return MimeType
}
