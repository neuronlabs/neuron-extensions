package authentication

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/server"

	"github.com/neuronlabs/neuron-extensions/codec/json"
	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
	"github.com/neuronlabs/neuron-extensions/server/http/middleware"
)

// API is an API for the accounts operations.
type API struct {
	Options       *Options
	Authenticator auth.Authenticator
	Tokener       auth.Tokener
	Endpoints     []*server.Endpoint

	// model is account model structure.
	model          *mapping.ModelStruct
	defaultHandler *DefaultHandler

	serverOptions server.Options
}

// New creates account API.
func New(options ...Option) (*API, error) {
	a := &API{
		Options:        defaultOptions(),
		defaultHandler: &DefaultHandler{},
	}
	for _, option := range options {
		option(a.Options)
	}
	if a.Options.AccountModel == nil {
		return nil, errors.Wrap(auth.ErrAccountModelNotDefined, "provided no account model for the account service")
	}
	if a.Options.PathPrefix != "" {
		if _, err := url.Parse(a.Options.PathPrefix); err != nil {
			return nil, errors.Wrap(server.ErrServer, "provided invalid path prefix for the authentication module")
		}
	}
	a.defaultHandler.Account = a.Options.AccountModel
	return a, nil
}

// InitializeAPI implements server/http.API interface.
func (a *API) InitializeAPI(options server.Options) error {
	a.serverOptions = options
	// Check authenticator.
	a.Authenticator = options.Authenticator
	if a.Authenticator == nil {
		return errors.Wrap(auth.ErrInitialization, "provided nil authenticator for the service")
	}
	a.Tokener = options.Tokener
	if a.Tokener == nil {
		return errors.Wrap(auth.ErrInitialization, "provided nil tokener for the service")
	}
	mStruct, err := a.serverOptions.Controller.ModelStruct(a.Options.AccountModel)
	if err != nil {
		return err
	}
	a.model = mStruct

	// Initialize default handler.
	if err := a.defaultHandler.Initialize(options.Controller); err != nil {
		return err
	}

	// Initialize handler if needed.
	if initializer, ok := a.Options.AccountHandler.(interface {
		Initialize(c *controller.Controller) error
	}); ok {
		if err := initializer.Initialize(options.Controller); err != nil {
			return err
		}
	}

	return nil
}

// GetEndpoints implements server.EndpointsGetter.
func (a *API) GetEndpoints() []*server.Endpoint {
	return a.Endpoints
}

// SetRoutes implements http server API interface.
func (a *API) SetRoutes(router *httprouter.Router) error {
	prefix := a.Options.PathPrefix
	if prefix == "" {
		prefix = "/"
	}

	// Register endpoint.
	middlewares := server.MiddlewareChain{middleware.StoreServerOptions(&a.serverOptions)}
	middlewares = append(middlewares, a.Options.Middlewares...)
	middlewares = append(middlewares, a.Options.RegisterMiddlewares...)
	router.POST(fmt.Sprintf("%s/register", prefix), httputil.Wrap(middlewares.
		Handle(http.HandlerFunc(a.handleRegisterAccount))))

	// Login endpoint.
	middlewares = server.MiddlewareChain{middleware.StoreServerOptions(&a.serverOptions)}
	middlewares = append(middlewares, a.Options.Middlewares...)
	middlewares = append(middlewares, a.Options.LoginMiddlewares...)
	router.POST(fmt.Sprintf("%s/login", prefix), httputil.Wrap(middlewares.Handle(http.HandlerFunc(a.handleLoginEndpoint))))

	// Refresh Token endpoint.
	middlewares = server.MiddlewareChain{middleware.StoreServerOptions(&a.serverOptions)}
	middlewares = append(middlewares, a.Options.Middlewares...)
	middlewares = append(middlewares, a.Options.RefreshTokenMiddlewares...)
	router.POST(fmt.Sprintf("%s/refresh", prefix), httputil.Wrap(middlewares.Handle(http.HandlerFunc(a.handleRefreshToken))))

	// Logout endpoint.
	middlewares = server.MiddlewareChain{middleware.StoreServerOptions(&a.serverOptions)}
	middlewares = append(middlewares, a.Options.Middlewares...)
	middlewares = append(middlewares, a.Options.LogoutMiddlewares...)
	router.POST(fmt.Sprintf("%s/logout", prefix), httputil.Wrap(middlewares.Handle(http.HandlerFunc(a.handleLogout))))

	return nil
}

func (a *API) marshalErrors(rw http.ResponseWriter, status int, err error) {
	httpErrors := httputil.MapError(err)
	a.setContentType(rw)
	// If no status is defined - set default from the errors.
	if status == 0 {
		status = codec.MultiError(httpErrors).Status()
	}
	// Write status to the header.
	rw.WriteHeader(status)
	// Marshal errors into response writer.
	marshalError := json.GetCodec(a.serverOptions.Controller).MarshalErrors(rw, httpErrors...)
	if err != nil {
		log.Errorf("Marshaling errors: '%v' failed: %v", httpErrors, marshalError)
	}
}

func (a *API) setContentType(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", json.MimeType)
}
