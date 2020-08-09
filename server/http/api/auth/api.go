package auth

import (
	"github.com/julienschmidt/httprouter"

	"github.com/neuronlabs/neuron/server"
)

// API is an API for the accounts operations.
type API struct {
	Options       *Options
	Endpoints     []*server.Endpoint
	serverOptions server.Options
}

// New creates account API.
func New(options ...Option) *API {
	a := &API{
		Options: &Options{},
	}
	for _, option := range options {
		option(a.Options)
	}
	return a
}

// InitializeAPI implements server/http.API interface.
func (a *API) InitializeAPI(options server.Options) error {
	a.serverOptions = options
	return nil
}

// GetEndpoints implements server.EndpointsGetter.
func (a *API) GetEndpoints() []*server.Endpoint {
	return a.Endpoints
}

// SetRoutes sets the router
func (a *API) SetRoutes(router *httprouter.Router) error {
	return nil
}
