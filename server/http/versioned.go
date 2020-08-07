package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/neuronlabs/neuron-extensions/server/http/log"
	"github.com/neuronlabs/neuron/server"
)

// Compile time check if VersionedServer implements server.VersionedServer.
var _ server.VersionedServer = &VersionedServer{}

// VersionedServer is a server.VersionedServer implementation for the http Server with multiple API versions.
type VersionedServer struct {
	Router    *httprouter.Router
	Endpoints []*server.Endpoint
	Options   *Options

	server http.Server
}

// NewVersioned creates new versioned API.
func NewVersioned(options ...Option) *VersionedServer {
	v := &VersionedServer{
		Options: newOptions(),
		server:  http.Server{},
		Router:  httprouter.New(),
	}

	for _, option := range options {
		option(v.Options)
	}
	v.setOptions()
	return v
}

// InitializeVersion initializes all API stored for given 'version'.
func (v *VersionedServer) InitializeVersion(version string, options server.Options) error {
	apis, ok := v.Options.VersionedAPIs[version]
	if !ok {
		return nil
	}
	for _, api := range apis {
		if err := api.InitializeAPI(options); err != nil {
			return err
		}
		if err := api.SetRoutes(v.Router); err != nil {
			return err
		}
		v.Endpoints = append(v.Endpoints, api.GetEndpoints()...)
	}
	return nil
}

// GetEndpoints implements server.VersionedServer.
func (v *VersionedServer) GetEndpoints() []*server.Endpoint {
	return v.Endpoints
}

// Serve implements server.VersionedServer.
func (v *VersionedServer) Serve() error {
	v.server.Handler = v.Router
	log.Infof("Listening and serve at: %s:%d", v.Options.Hostname, v.Options.Port)
	return v.server.ListenAndServe()
}

// Shutdown gently shutdown the server connection.
func (v *VersionedServer) Shutdown(ctx context.Context) error {
	if err := v.server.Shutdown(ctx); err != nil {
		log.Errorf("HTTP server shutdown failed: %v", err)
		return err
	}
	return nil
}

func (v *VersionedServer) setOptions() {
	if v.Options.Port == 0 {
		v.Options.Port = 80
	}
	v.server.Addr = fmt.Sprintf("%s:%d", v.Options.Hostname, v.Options.Port)
	v.server.TLSConfig = v.Options.TLSConfig
}
