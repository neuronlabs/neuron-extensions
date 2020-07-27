package json

import (
	"encoding/json"
	"io"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
)

var c *jsonCodec

const MimeType = "application/json"

func init() {
	c = &jsonCodec{Mime: MimeType}
	codec.RegisterCodec(MimeType, c)
}

var _ codec.Codec = &jsonCodec{}

type jsonCodec struct {
	Mime string
	c    *controller.Controller
}

// Initialize implements core.Initializer interface.
func (j *jsonCodec) Initialize(c *controller.Controller) error {
	j.c = c
	return nil
}

// MarshalErrors implements codec.Codec interface.
func (j jsonCodec) MarshalErrors(w io.Writer, errs ...*codec.Error) error {
	data, err := json.Marshal(errs)
	if err != nil {
		return errors.Newf(codec.ClassMarshal, "marshaling errors failed: %v", err)
	}
	_, err = w.Write(data)
	if err != nil {
		return errors.Newf(codec.ClassMarshal, "writing marshaled errors failed: %v", err)
	}
	return nil
}

// UnmarshalErrors implements codec.Codec interface.
func (j jsonCodec) UnmarshalErrors(r io.Reader) (codec.MultiError, error) {
	errs := []*codec.Error{}
	if err := json.NewDecoder(r).Decode(&errs); err != nil {
		return nil, errors.NewDetf(codec.ClassUnmarshal, "unmarshal errors failed: %v", err)
	}
	return errs, nil
}

// MimeType implements codec.Codec interface.
func (j jsonCodec) MimeType() string {
	return j.Mime
}
