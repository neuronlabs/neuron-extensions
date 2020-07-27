package jsonapi

import (
	"net/http"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/db"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
	"github.com/neuronlabs/neuron/server"

	"github.com/neuronlabs/neuron-extensions/codec/jsonapi"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
)

// HandleGet handles json:api get endpoint for the 'model'. Panics if the model is not mapped for given API controller.
func (a *API) HandleGet(model mapping.Model) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		a.handleGet(a.Controller.MustModelStruct(model))(rw, req)
	}
}

func (a *API) handleGet(model *mapping.ModelStruct) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		s, err := a.createGetScope(req, model)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		params := &QueryParams{
			Context: req.Context(),
			Scope:   s,
			DB:      a.DB,
		}

		// Get and apply hook functions.
		for _, hook := range hooks.getPreHooks(model, query.Get) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		_, err = db.Get(params.Context, params.DB, params.Scope)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		for _, hook := range hooks.getPostHooks(model, query.Get) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		linkType := codec.ResourceLink
		// but if the config doesn't allow that - set 'jsonapi.NoLink'
		if !a.MarshalLinks {
			linkType = codec.NoLink
		}

		// TODO: add relations.
		payload := codec.Payload{
			ModelStruct:       s.ModelStruct,
			Data:              s.Models,
			FieldSets:         s.FieldSets,
			IncludedRelations: s.IncludedRelations,
			MarshalLinks: &codec.LinkOptions{
				Type:       linkType,
				BaseURL:    a.BasePath,
				RootID:     httputil.CtxMustGetID(params.Context),
				Collection: model.Collection(),
			},
			MarshalSingularFormat: true,
		}
		a.marshalPayload(rw, &payload, http.StatusOK)
	}
}

func (a *API) createGetScope(req *http.Request, mStruct *mapping.ModelStruct) (*query.Scope, error) {
	id := httputil.CtxMustGetID(req.Context())
	if id == "" {
		log.Errorf("ID value stored in the context is empty.")
		return nil, errors.NewDet(server.ClassURIParameter, "invalid 'id' url parameter").
			WithDetail("Provided empty ID in query url")
	}

	// Create new model and set it's primary key from the url parameter.
	model := mapping.NewModel(mStruct)
	if err := model.SetPrimaryKeyStringValue(id); err != nil {
		log.Debugf("[GET][%s] Invalid URL id value: '%s': '%v'", mStruct.Collection(), id, err)
		return nil, errors.NewDet(server.ClassURIParameter, "invalid query id parameter")
	}

	// Create a query scope and parse url parameters.
	s := query.NewScope(mStruct)

	// Set primary key filter for given model.
	s.Filter(filter.New(mStruct.Primary(), filter.OpEqual, model.GetPrimaryKeyValue()))

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
