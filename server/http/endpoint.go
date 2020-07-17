package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/orm"
)

type Middleware func(httprouter.Handle) httprouter.Handle

type EndpointInitializer func(c *controller.Controller, db orm.DB) *Endpoint

// Endpoint is the http server endpoint based on the given mime type and path.
type Endpoint struct {
	Method      string
	MimeType    string
	Middlewares []Middleware
	Handle      httprouter.Handle
	Path        string
}

func (e *Endpoint) handle(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if len(e.Middlewares) == 0 {
		e.Handle(rw, req, params)
		return
	}
	wrapped := e.Handle
	for i := len(e.Middlewares); i >= 0; i++ {
		wrapped = e.Middlewares[i](wrapped)
	}
	wrapped(rw, req, params)
}

func EndpointsHandler(endpoints map[string]*Endpoint) httprouter.Handle {
	return func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		accepts := parseQVHeader(req.Header, "Accept")
		for _, accept := range accepts {
			endpoint, ok := endpoints[accept.Value]
			if !ok {
				continue
			}
			endpoint.handle(rw, req, params)
			return
		}

	}
}
