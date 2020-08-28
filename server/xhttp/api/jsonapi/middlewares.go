package jsonapi

import (
	"fmt"
	"net/http"

	"github.com/neuronlabs/neuron-extensions/codec/jsonapi"
	"github.com/neuronlabs/neuron-extensions/server/xhttp/httputil"
	"github.com/neuronlabs/neuron-extensions/server/xhttp/log"
	"github.com/neuronlabs/neuron/core"
)

// MidAccept creates a middleware that requires provided accept
func MidAccept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		parsed := httputil.ParseAcceptHeader(req.Header)
		for _, qv := range parsed {
			if qv.Value == jsonapi.MimeType {
				next.ServeHTTP(rw, req)
				return
			}
		}

		rw.WriteHeader(http.StatusNotAcceptable)
		c, ok := core.CtxGetController(req.Context())
		if !ok {
			return
		}
		err := httputil.ErrUnsupportedHeader()
		err.Detail = fmt.Sprintf("header Accept doesn't contain '%s' mime type", jsonapi.MimeType)
		if err := jsonapi.GetCodec(c).MarshalErrors(rw, err); err != nil {
			log.Errorf("Marshaling error failed: %v", err)
		}
	})
}

// MidAccept creates a middleware that requires provided accept
func MidContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ct := req.Header.Get("Content-Type")
		if ct == jsonapi.MimeType {
			next.ServeHTTP(rw, req)
			return
		}
		rw.WriteHeader(http.StatusUnsupportedMediaType)
	})
}
