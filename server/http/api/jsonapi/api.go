package jsonapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/julienschmidt/httprouter"

	httpServer "github.com/neuronlabs/neuron-extensions/server/http"
	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/middleware"
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/db"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/server"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-extensions/codec/jsonapi"
)

// Compile time check if API implements httpServer.API.
var _ httpServer.API = &API{}

// API is the neuron handler that implements https://jsonapi.org server routes for neuron models.
type API struct {
	// BasePath is the basic path used by the handler.
	BasePath string
	// DefaultPageSize defines default PageSize for the list endpoints.
	DefaultPageSize int
	// NoContentOnCreate allows to set the flag for the models with client generated id to return no content.
	NoContentOnCreate bool
	// StrictFieldsMode defines if the during unmarshal process the query should strictly check
	// if all the fields are well known to given model.
	StrictUnmarshal bool
	// IncludeNestedLimit is a maximum value for nested includes (i.e. IncludeNestedLimit = 1
	// allows ?include=posts.comments but does not allow ?include=posts.comments.author)
	IncludeNestedLimit int
	// FilterValueLimit is a maximum length of the filter values
	FilterValueLimit int
	// MarshalLinks is the default behavior for marshaling the resource links into the handler responses.
	MarshalLinks bool
	// Middlewares are global middlewares added to each endpoint in the API.
	Middlewares middleware.Chain

	// Server options.
	Authorizer    auth.Authorizer
	Authenticator auth.Authenticator
	DB            db.DB
	Controller    *controller.Controller

	modelEndpoints   map[string][]*Endpoint
	defaultEndpoints map[string]struct{}
	hooks            *hooksStore
}

// New creates new jsonapi API API for the Default Controller.
func New() *API {
	return &API{
		hooks:            newHooksStore(),
		modelEndpoints:   map[string][]*Endpoint{},
		defaultEndpoints: map[string]struct{}{},
	}
}

// Init implements httpServer.API interface.
func (a *API) Init(options server.Options) error {
	a.Controller = options.Controller
	a.DB = options.DB
	a.Authorizer = options.Authorizer
	a.Authenticator = options.Authenticator

	// Set codec as default in the context.
	a.Middlewares = append(middleware.Chain{middleware.WithCodec(jsonapi.MimeType)}, a.Middlewares...)

	// check if the base path has absolute value - if not add the leading slash to the BasePath.
	if !path.IsAbs(a.BasePath) {
		a.BasePath = "/" + a.BasePath
	}
	return nil
}

// Set implements RoutesSetter.
func (a *API) SetRoutes(router *httprouter.Router) error {
	for collection, endpoints := range a.modelEndpoints {
		model, ok := a.Controller.ModelMap.GetByCollection(collection)
		if !ok {
			return errors.NewDetf(mapping.ClassModelNotFound, "model not found for collection: '%s'", collection)
		}

		// Set the default endpoints if the model have them defined.
		if _, ok = a.defaultEndpoints[model.Collection()]; ok {
			a.setModelsDefaultEndpoints(model)
			endpoints = a.modelEndpoints[collection]
		}

		for _, endpoint := range endpoints {
			// Apply global middlewares.
			endpoint.Middlewares = append(a.Middlewares, endpoint.Middlewares...)
			if endpoint.Handler == nil {
				if err := a.setEndpointHandlerFunc(model, endpoint); err != nil {
					return err
				}
			}
			switch endpoint.Method {
			case query.Insert:
				router.POST(fmt.Sprintf("%s/%s", a.BasePath, model.Collection()), httputil.Wrap(endpoint.Middlewares.Handle(endpoint.Handler)))
			case query.InsertRelationship:
				relation, err := a.getRelation(endpoint, model)
				if err != nil {
					return err
				}
				router.POST(fmt.Sprintf("%s/%s/%s", a.BasePath, model.Collection(), relation.NeuronName()), httputil.Wrap(endpoint.Middlewares.Handle(endpoint.Handler)))
			case query.Delete:
				router.DELETE(fmt.Sprintf("%s/%s/:id", a.BasePath, model.Collection()), httputil.Wrap(endpoint.Middlewares.Handle(endpoint.Handler)))
			case query.DeleteRelationship:
				relation, err := a.getRelation(endpoint, model)
				if err != nil {
					return err
				}
				router.DELETE(fmt.Sprintf("%s/%s/%s", a.BasePath, model.Collection(), relation.NeuronName()), httputil.Wrap(endpoint.Middlewares.Handle(endpoint.Handler)))
			case query.Get:
				router.GET(fmt.Sprintf("%s/%s/:id", a.BasePath, model.Collection()), httputil.Wrap(endpoint.Middlewares.Handle(endpoint.Handler)))
			case query.GetRelated:
				relation, err := a.getRelation(endpoint, model)
				if err != nil {
					return err
				}
				router.GET(fmt.Sprintf("%s/%s/:id/%s", a.BasePath, model.Collection(), relation.NeuronName()), httputil.Wrap(endpoint.Middlewares.Handle(endpoint.Handler)))
			case query.GetRelationship:
				relation, err := a.getRelation(endpoint, model)
				if err != nil {
					return err
				}
				router.GET(fmt.Sprintf("%s/%s/:id/relationships/%s", a.BasePath, model.Collection(), relation.NeuronName()), httputil.Wrap(endpoint.Middlewares.Handle(endpoint.Handler)))
			case query.List:
				router.GET(fmt.Sprintf("%s/%s", a.BasePath, model.Collection()), httputil.Wrap(endpoint.Middlewares.Handle(endpoint.Handler)))
			case query.Update:
				router.PATCH(fmt.Sprintf("%s/%s/:id", a.BasePath, model.Collection()), httputil.Wrap(endpoint.Middlewares.Handle(endpoint.Handler)))
			case query.UpdateRelationship:
				relation, err := a.getRelation(endpoint, model)
				if err != nil {
					return err
				}
				router.PATCH(fmt.Sprintf("%s/%s/:id/relationships/%s", a.BasePath, model.Collection(), relation.NeuronName()), httputil.Wrap(endpoint.Middlewares.Handle(endpoint.Handler)))
			}
		}
	}
	return nil
}

func (a *API) getRelation(endpoint *Endpoint, model *mapping.ModelStruct) (*mapping.StructField, error) {
	if endpoint.Relation != "" {
		relation, ok := model.RelationByName(endpoint.Relation)
		if !ok {
			return nil, errors.NewDetf(mapping.ClassInvalidRelationField, "provided models:'%s' field:'%s' is not a valid relation", endpoint.Relation)
		}
		return relation, nil
	}
	return nil, errors.NewDetf(server.ClassEndpoint, "missing relation definition in the endpoint: '%v' for model: '%s'", endpoint.Method, model)
}

func (a *API) basePath() string {
	if a.BasePath == "" {
		return "/"
	}
	return a.BasePath
}

func (a *API) baseModelPath(mStruct *mapping.ModelStruct) string {
	return path.Join("/", a.BasePath, mStruct.Collection())
}

func (a *API) writeContentType(rw http.ResponseWriter) {
	rw.Header().Add("Content-Type", jsonapi.MimeType)
}

func (a *API) jsonapiUnmarshalOptions() *codec.UnmarshalOptions {
	return &codec.UnmarshalOptions{StrictUnmarshal: a.StrictUnmarshal}
}

func (a *API) marshalErrors(rw http.ResponseWriter, status int, errs ...*codec.Error) {
	a.writeContentType(rw)
	// If no status is defined - set default from the errors.
	if status == 0 {
		status = codec.MultiError(errs).Status()
	}
	// Write status to the header.
	rw.WriteHeader(status)
	// Marshal errors into response writer.
	err := jsonapi.Codec().MarshalErrors(rw, errs...)
	if err != nil {
		log.Errorf("Marshaling errors: '%v' failed: %v", errs, err)
	}
}

func (a *API) marshalPayload(rw http.ResponseWriter, payload *codec.Payload, status int) {
	a.writeContentType(rw)
	buf := &bytes.Buffer{}
	payloadMarshaler := jsonapi.Codec().(codec.PayloadMarshaler)
	if err := payloadMarshaler.MarshalPayload(buf, payload); err != nil {
		rw.WriteHeader(500)
		err := jsonapi.Codec().MarshalErrors(rw, httputil.ErrInternalError())
		if err != nil {
			switch err {
			case io.ErrShortWrite, io.ErrClosedPipe:
				log.Debug2f("An error occurred while writing api errors: %v", err)
			default:
				log.Errorf("Marshaling error failed: %v", err)
			}
		}
		return
	}
	rw.WriteHeader(status)
	if _, err := rw.Write(buf.Bytes()); err != nil {
		log.Errorf("Writing to response writer failed: %v", err)
	}
}

func (a *API) createListScope(model *mapping.ModelStruct, req *http.Request) (*query.Scope, error) {
	// Create a query scope and parse url parameters.
	s := query.NewScope(model)
	// Get jsonapi codec ans parse query parameters.
	parser, ok := jsonapi.Codec().(codec.ParameterParser)
	if !ok {
		log.Errorf("jsonapi codec doesn't implement ParameterParser")
		return nil, errors.NewDet(codec.ClassInternal, "jsonapi codec doesn't implement ParameterParser")
	}

	parameters := query.MakeParameters(req.URL.Query())
	if err := parser.ParseParameters(a.Controller, s, parameters); err != nil {
		return nil, err
	}
	return s, nil
}

func (a *API) getBasePath(basePath string) string {
	if basePath == "" {
		basePath = a.BasePath
	}
	return basePath
}
