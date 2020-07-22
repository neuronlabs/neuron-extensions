package jsonapi

import (
	"fmt"
	"net/http"

	"github.com/neuronlabs/neuron-plugins/codec/jsonapi"
	"github.com/neuronlabs/neuron-plugins/server/http/httputil"
	"github.com/neuronlabs/neuron-plugins/server/http/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/orm"
	"github.com/neuronlabs/neuron/query"
)

// HandleDeleteRelationship handles json:api delete relationship endpoint for the 'model'.
// Panics if the model is not mapped for given API controller or the relation doesn't exists.
func (a *API) HandleDeleteRelationship(model mapping.Model, relationName string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		mStruct := a.Controller.MustModelStruct(model)
		relation, ok := mStruct.RelationByName(relationName)
		if !ok {
			panic(fmt.Sprintf("no relation: '%s' found for the model: '%s'", relationName, mStruct.Type().Name()))
		}
		a.handleDeleteRelationship(mStruct, relation)(rw, req)
	}
}

func (a *API) handleDeleteRelationship(mStruct *mapping.ModelStruct, relation *mapping.StructField) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Get the id from the url.
		id := httputil.CtxMustGetID(req.Context())
		if id == "" {
			log.Debugf("[DELETE-RELATIONSHIP][%s] Empty id params", mStruct.Collection())
			err := httputil.ErrBadRequest()
			err.Detail = "Provided empty 'id' in url"
			a.marshalErrors(rw, 0, err)
			return
		}

		model := mapping.NewModel(mStruct)
		if err := model.SetPrimaryKeyStringValue(id); err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		relations, err := jsonapi.Codec().UnmarshalModels(req.Body, relation.Relationship().Struct(), a.jsonapiUnmarshalOptions())
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		for _, relation := range relations {
			if relation.IsPrimaryKeyZero() {
				err := httputil.ErrInvalidJSONFieldValue()
				err.Detail = "one of provided relationships doesn't have it's primary key value stored"
				a.marshalErrors(rw, 0, err)
				return
			}
		}

		if len(relations) == 0 {
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		s := query.NewScope(mStruct)
		s.FieldSet = mapping.FieldSet{mStruct.Primary()}
		if err = s.Filter(query.NewFilterField(mStruct.Primary(), query.OpEqual, model.GetPrimaryKeyValue())); err != nil {
			a.marshalErrors(rw, 500, httputil.ErrInternalError())
			return
		}

		// Include relation values.
		if err = s.Include(relation, relation.Relationship().Struct().Primary()); err != nil {
			a.marshalErrors(rw, 500, httputil.ErrInternalError())
			return
		}

		params := &QueryParams{
			Context:   req.Context(),
			DB:        a.DB,
			Scope:     s,
			Relations: []*mapping.StructField{relation},
		}

		// Get and apply pre hook functions.
		for _, hook := range hooks.getPreHooksRelation(mStruct, query.DeleteRelationship, relation.Name()) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		// Check if the model with provided id exists.
		model, err = orm.Get(params.Context, params.DB, params.Scope)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		var relationsToSet []mapping.Model
		switch relation.Kind() {
		case mapping.KindRelationshipMultiple:
			mr, ok := model.(mapping.MultiRelationer)
			if !ok {
				a.marshalErrors(rw, 500, httputil.ErrInternalError())
				return
			}
			models, err := mr.GetRelationModels(relation)
			if err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}

			for _, relationModel := range models {
				var isToDelete bool
				for _, toDeleteModel := range relations {
					if relationModel.GetPrimaryKeyHashableValue() == toDeleteModel.GetPrimaryKeyHashableValue() {
						isToDelete = true
						break
					}
				}
				if !isToDelete {
					relationsToSet = append(relationsToSet, relationModel)
				}
			}
		case mapping.KindRelationshipSingle:
			sr, ok := model.(mapping.SingleRelationer)
			if !ok {
				a.marshalErrors(rw, 500, httputil.ErrInternalError())
				return
			}
			relationModel, err := sr.GetRelationModel(relation)
			if err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}

			if relationModel == nil {
				rw.WriteHeader(http.StatusNoContent)
				return
			}
			var alreadySet bool
			for _, relation := range relations {
				if relation.GetPrimaryKeyHashableValue() == relationModel.GetPrimaryKeyHashableValue() {
					alreadySet = true
					break
				}
			}
			if !alreadySet {
				rw.WriteHeader(http.StatusNoContent)
				return
			}
		}

		if len(relationsToSet) == 0 {
			_, err = orm.RemoveRelations(params.Context, params.DB, params.Scope, relation)
		} else {
			err = orm.SetRelations(params.Context, params.DB, params.Scope, relation, relationsToSet...)
		}
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		// Get and apply pre hook functions.
		for _, hook := range hooks.getPostHooksRelation(mStruct, query.DeleteRelationship, relation.Name()) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}
		rw.WriteHeader(http.StatusNoContent)
	}
}
