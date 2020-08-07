package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/neuronlabs/neuron-extensions/server/http/log"
	"github.com/neuronlabs/neuron/server"
)

var _ server.Server = &Server{}

// API is an interface used for the server API's.
type API interface {
	server.EndpointsGetter
	InitializeAPI(options server.Options) error
	SetRoutes(router *httprouter.Router) error
}

// Server is an http server implementation. It implements neuron/server.Server interface.
type Server struct {
	Options   *Options
	Router    *httprouter.Router
	Endpoints []*server.Endpoint

	serverOptions server.Options
	server        http.Server
}

// New creates new server.
func New(options ...Option) *Server {
	s := &Server{
		Options: newOptions(),
		server:  http.Server{},
		Router:  httprouter.New(),
	}
	for _, option := range options {
		option(s.Options)
	}
	return s
}

func newOptions() *Options {
	return &Options{VersionedAPIs: map[string][]API{}}
}

// GetEndpoints gets all stored server endpoints.
func (s *Server) GetEndpoints() []*server.Endpoint {
	return s.Endpoints
}

// Initialize initializes server with provided options.
func (s *Server) Initialize(options server.Options) error {
	// Initialize all endpoints.
	s.serverOptions = options
	s.setOptions()

	for _, api := range s.Options.APIs {
		if err := api.InitializeAPI(options); err != nil {
			return err
		}
		if err := api.SetRoutes(s.Router); err != nil {
			return err
		}
		s.Endpoints = append(s.Endpoints, api.GetEndpoints()...)
	}

	for _, apis := range s.Options.VersionedAPIs {
		for _, api := range apis {
			if err := api.InitializeAPI(options); err != nil {
				return err
			}
			if err := api.SetRoutes(s.Router); err != nil {
				return err
			}
			s.Endpoints = append(s.Endpoints, api.GetEndpoints()...)
		}
	}
	return nil
}

// Serve serves all routes stored in given server.
func (s *Server) Serve() error {
	s.server.Handler = s.Router
	log.Infof("Listening and serve at: %s:%d", s.Options.Hostname, s.Options.Port)
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

func (s *Server) setOptions() {
	if s.Options.Port == 0 {
		s.Options.Port = 80
	}
	s.server.Addr = fmt.Sprintf("%s:%d", s.Options.Hostname, s.Options.Port)
	s.server.TLSConfig = s.Options.TLSConfig
}
