package http

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/server"
)

var _ server.Server = &Server{}

type Server struct {
	// Endpoints provides the mapping between the path - mime type and given endpoint handler
	// map[path]map[method]map[mime]
	Endpoints            map[string]map[string]map[string]*Endpoint
	EndpointInitializers []EndpointInitializer
	Options              server.Options
	server               http.Server
	router               *httprouter.Router
}

func New() *Server {
	return &Server{
		Endpoints: map[string]map[string]map[string]*Endpoint{},
		server:    http.Server{},
		router:    httprouter.New(),
	}
}

func (s *Server) Serve() error {
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Server) Initialize(options *server.Options) error {
	// Initialize all endpoints.

	for _, initializer := range s.EndpointInitializers {
		endpoint := initializer(c, db)
		methodsMap, ok := s.Endpoints[endpoint.Path]
		if !ok {
			methodsMap = map[string]map[string]*Endpoint{}
			s.Endpoints[endpoint.Path] = methodsMap
		}
		mimeMap, ok := methodsMap[endpoint.Method]
		if !ok {
			mimeMap = map[string]*Endpoint{}
			methodsMap[endpoint.Method] = mimeMap
		}

		_, ok = mimeMap[endpoint.MimeType]
		if ok {
			return errors.Newf(server.ClassDuplicatedEndpoint, "endpoint: %s %s %s is already registered for the server",
				endpoint.Method, endpoint.Path, endpoint.MimeType)
		}
		mimeMap[endpoint.MimeType] = endpoint

	}
	return nil
}
