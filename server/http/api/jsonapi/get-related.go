package jsonapi

import (
	"fmt"
	"net/http"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/db"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
)

// HandleGetRelated handles json:api get related endpoint for the 'model'.
// Panics if the model is not mapped for given API controller or relationName is not found.
func (a *API) HandleGetRelated(model mapping.Model, relationName string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		mStruct := a.Controller.MustModelStruct(model)
		relation, ok := mStruct.RelationByName(relationName)
		if !ok {
			panic(fmt.Sprintf("no relation: '%s' found for the model: '%s'", relationName, mStruct.Type().Name()))
		}
		a.handleGetRelated(mStruct, relation)(rw, req)
	}
}

func (a *API) handleGetRelated(mStruct *mapping.ModelStruct, field *mapping.StructField) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		// Check the URL 'id' value.
		id := httputil.CtxMustGetID(ctx)
		if id == "" {
			log.Debugf("[GET-RELATED][%s] Empty id params", mStruct.Collection())
			err := httputil.ErrBadRequest()
			err.Detail = "Provided empty 'id' in url"
			a.marshalErrors(rw, 0, err)
			return
		}

		model := mapping.NewModel(mStruct)
		err := model.SetPrimaryKeyStringValue(id)
		if err != nil {
			log.Debugf("[GET-RELATED][%s] Invalid URL id value: '%s': '%v'", mStruct.Collection(), id, err)
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		// Set preset filters.
		s := query.NewScope(mStruct)
		// Set the primary field value.
		if s.Filter(filter.New(mStruct.Primary(), filter.OpEqual, model.GetPrimaryKeyValue())); err != nil {
			log.Errorf("[GET-RELATED][%s][%s] Adding param primary filter with value: '%s' failed: %v", mStruct.Collection(), field.NeuronName(), id, err)
			a.marshalErrors(rw, 0, httputil.ErrInternalError())
			return
		}

		// Include relation.
		if err = s.Include(field); err != nil {
			log.Errorf("[GET-RELATED][%s][%s] Setting related field into fieldset failed: %v", mStruct.Collection(), field.NeuronName(), err)
			a.marshalErrors(rw, 0, httputil.ErrInternalError())
			return
		}

		// Create Parameters.
		params := &QueryParams{
			Context:   ctx,
			DB:        a.DB,
			Scope:     s,
			Relations: mapping.FieldSet{field},
		}

		for _, hook := range hooks.getPreHooksRelation(mStruct, query.GetRelated, field.Name()) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		model, err = db.Get(params.Context, params.DB, params.Scope)
		if err != nil {
			log.Errorf("[GET-RELATED][%s][%s] Setting related field into fieldset failed: %v", mStruct.Collection(), field.NeuronName(), err)
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		//
		for _, hook := range hooks.getPostHooksRelation(mStruct, query.GetRelated, field.Name()) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		linkType := codec.RelatedLink
		// but if the config doesn't allow that - set 'codec.NoLink'
		if !a.MarshalLinks {
			linkType = codec.NoLink
		}

		var models []mapping.Model
		if field.Relationship().IsToMany() {
			mr, ok := model.(mapping.MultiRelationer)
			if !ok {
				log.Errorf("Model: '%s' is not MultiRelationer", mStruct.Collection())
				a.marshalErrors(rw, 500, httputil.ErrInternalError())
				return
			}
			models, err = mr.GetRelationModels(field)
			if err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		} else {
			sr, ok := model.(mapping.SingleRelationer)
			if !ok {
				log.Errorf("Model: '%s' is not MultiRelationer", mStruct.Collection())
				a.marshalErrors(rw, 500, httputil.ErrInternalError())
				return
			}
			relationModel, err := sr.GetRelationModel(field)
			if err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
			models = append(models, relationModel)
		}
		// TODO: add relations.
		payload := codec.Payload{
			ModelStruct: field.Relationship().Struct(),
			Data:        models,
			FieldSets:   []mapping.FieldSet{field.Relationship().Struct().Fields()},
			MarshalLinks: &codec.LinkOptions{
				Type:          linkType,
				BaseURL:       a.BasePath,
				RootID:        id,
				Collection:    mStruct.Collection(),
				RelationField: field.NeuronName(),
			},
			MarshalSingularFormat: false,
		}
		a.marshalPayload(rw, &payload, http.StatusOK)
	}
}
