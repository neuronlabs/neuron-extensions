package jsonapi

import (
	"net/http"

	"github.com/neuronlabs/neuron/db"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
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

		s := query.NewScope(mStruct)
		if s.Filter(filter.New(mStruct.Primary(), filter.OpEqual, model.GetPrimaryKeyValue())); err != nil {
			// this should not occur - primary field's model must match scope's model.
			log.Errorf("[DELETE][%s] Adding param primary filter with value: '%s' failed: %v", mStruct.Collection(), id, err)
			a.marshalErrors(rw, 0, httputil.ErrInternalError())
			return
		}

		// Set endpoint parameters.
		params := &QueryParams{
			Context: req.Context(),
			DB:      a.DB,
			Scope:   s,
		}

		// Get and apply hook functions.
		for _, hook := range hooks.getPreHooks(mStruct, query.Delete) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		if _, err = db.Delete(params.Context, params.DB, params.Scope); err != nil {
			log.Debugf("[DELETE][SCOPE][%s] Delete %s/%s root scope failed: %v", s.ID, mStruct.Collection(), id, err)
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		// Get and apply post hook functions.
		for _, hook := range hooks.getPostHooks(mStruct, query.Delete) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		// Write no content status.
		rw.WriteHeader(http.StatusNoContent)
	}
}
