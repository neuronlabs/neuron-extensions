package http

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/neuronlabs/neuron-plugins/server/http/log"
	"github.com/neuronlabs/neuron/server"
)

var _ server.Server = &Server{}

// API is an interface used for the server API's.
type API interface {
	Init(options server.Options) error
	SetRoutes(router *httprouter.Router) error
}

// Server is an http server implementation. It implements neuron/server.Server interface.
type Server struct {
	Options server.Options
	server  http.Server
	Router  *httprouter.Router
	APIs    []API
}

// New creates new server.
func New() *Server {
	return &Server{
		server: http.Server{},
		Router: httprouter.New(),
	}
}

// SetAPI sets provided API on given server.
func (s *Server) SetAPI(a API) {
	s.APIs = append(s.APIs, a)
}

// Serve serves all routes stored in given server.
func (s *Server) Serve() error {
	s.server.Handler = s.Router
	return s.server.ListenAndServe()
}

// Shutdown gently shutdown the server connection.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		log.Errorf("HTTP server shutdown failed: %v", err)
		return err
	}
	return nil
}

// Initialize initializes server with provided options.
func (s *Server) Initialize(options server.Options) error {
	// Initialize all endpoints.
	s.Options = options
	for _, api := range s.APIs {
		if err := api.Init(options); err != nil {
			return err
		}
		if err := api.SetRoutes(s.Router); err != nil {
			return err
		}
	}
	return nil
}
