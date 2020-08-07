package jsonapi

import (
	"net/http"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/server"
)

// HandleDelete handles json:api delete endpoint for the 'model'. Panics if the model is not mapped for given API controller.
func (a *API) HandleDelete(model mapping.Model) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		a.handleDelete(a.Controller.MustModelStruct(model))(rw, req)
	}
}

func (a *API) handleDelete(mStruct *mapping.ModelStruct) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		id := httputil.CtxMustGetID(ctx)
		if id == "" {
			// if the function would not contain 'id' parameter.
			log.Debugf("[DELETE] Empty id params: %v", id)
			err := httputil.ErrInvalidQueryParameter()
			err.Detail = "Provided empty id in the query URL"
			a.marshalErrors(rw, 0, err)
			return
		}

		model := mapping.NewModel(mStruct)
		err := model.SetPrimaryKeyStringValue(id)
		if err != nil {
			log.Debugf("[DELETE][%s] Invalid URL id value: '%s': '%v'", mStruct.Collection(), id, err)
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		// Check if the primary key is not zero value.
		if model.IsPrimaryKeyZero() {
			err := httputil.ErrInvalidQueryParameter()
			err.Detail = "provided zero value primary key for the model"
			a.marshalErrors(rw, 0, err)
			return
		}
		// Create scope for the delete purpose.
		s := query.NewScope(mStruct, model)

		params := &server.Params{
			Ctx:           req.Context(),
			DB:            a.DB,
			Authenticator: a.Authenticator,
			Authorizer:    a.Authorizer,
		}

		modelHandler, hasModelHandler := a.handlers[mStruct]
		if hasModelHandler {
			beforeDeleter, ok := modelHandler.(server.BeforeDeleteHandler)
			if ok {
				if err = beforeDeleter.HandleBeforeDelete(params, s); err != nil {
					a.marshalErrors(rw, 0, httputil.MapError(err)...)
					return
				}
			}
		}

		deleteHandler, ok := modelHandler.(server.DeleteHandler)
		if !ok {
			deleteHandler = a.defaultHandler
		}

		result, err := deleteHandler.HandleDelete(ctx, *params, s)
		if err != nil {
			log.Debugf("[DELETE][SCOPE][%s] Delete %s/%s root scope failed: %v", s.ID, mStruct.Collection(), id, err)
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}
		if hasModelHandler {
			afterHandler, ok := modelHandler.(server.AfterDeleteHandler)
			if ok {
				if err = afterHandler.HandleAfterDelete(params, s, result); err != nil {
					a.marshalErrors(rw, 0, httputil.MapError(err)...)
					return
				}
			}
		}

		if result == nil || result.Meta == nil {
			// Write no content status.
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		a.marshalPayload(rw, result, http.StatusOK)
	}
}
