package auth

import (
	"net/http"

	"github.com/neuronlabs/neuron-extensions/codec/json"
)

func (a *API) createAccount() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		codec := json.GetCodec(a.serverOptions.Controller)
	}
}
