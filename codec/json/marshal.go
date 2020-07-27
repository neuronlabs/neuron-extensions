package json

import (
	"io"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/mapping"
)

func (j jsonCodec) MarshalModels(w io.Writer, modelStruct *mapping.ModelStruct, models []mapping.Model, options *codec.MarshalOptions) error {
	panic("implement me")
}
