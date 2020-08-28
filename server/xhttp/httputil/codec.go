package httputil

import (
	"context"

	"github.com/neuronlabs/neuron/codec"
)

type codecKey struct{}

var ctxCodecKey = &codecKey{}

// GetCodec gets the codec from the context.
func GetCodec(ctx context.Context) (codec.Codec, bool) {
	c, ok := ctx.Value(ctxCodecKey).(codec.Codec)
	return c, ok
}

// SetCodec sets the codec in the context.
func SetCodec(ctx context.Context, c codec.Codec) context.Context {
	return context.WithValue(ctx, ctxCodecKey, c)
}
