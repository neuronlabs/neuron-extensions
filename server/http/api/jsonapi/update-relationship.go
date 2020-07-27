package jsonapi

import (
	"fmt"
	"net/http"

	"github.com/neuronlabs/neuron-extensions/codec/jsonapi"
	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/db"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
)

// HandleUpdateRelationship handles json:api update relationship endpoint for the 'model'.
// Panics if the model is not mapped for given API controller or the relation doesn't exists.
func (a *API) HandleUpdateRelationship(model mapping.Model, relationName string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		mStruct := a.Controller.MustModelStruct(model)
		relation, ok := mStruct.RelationByName(relationName)
		if !ok {
			panic(fmt.Sprintf("no relation: '%s' found for the model: '%s'", relationName, mStruct.Type().Name()))
		}
		a.handleUpdateRelationship(mStruct, relation)(rw, req)
	}
}

func (a *API) handleUpdateRelationship(mStruct *mapping.ModelStruct, relation *mapping.StructField) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Get the id from the url.
		id := httputil.CtxMustGetID(req.Context())
		if id == "" {
			log.Debugf("[PATCH][%s] Empty id params", mStruct.Collection())
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

		var removeRelationship bool
		pu := jsonapi.Codec().(codec.PayloadUnmarshaler)
		payload, err := pu.UnmarshalPayload(req.Body, codec.UnmarshalOptions{
			StrictUnmarshal: a.StrictUnmarshal,
			ModelStruct:     relation.Relationship().Struct(),
		})
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if len(payload.Data) == 0 {
			removeRelationship = true
		}

		s := query.NewScope(mStruct, model)
		s.FieldSets = payload.FieldSets

		params := &QueryParams{
			Context:   req.Context(),
			DB:        a.DB,
			Scope:     s,
			Relations: nil,
		}

		// Get and apply pre hook functions.
		for _, hook := range hooks.getPreHooksRelation(mStruct, query.UpdateRelationship, relation.Name()) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		// Check if the model with provided id exists.
		existScope := query.NewScope(mStruct)
		if existScope.Filter(filter.New(mStruct.Primary(), filter.OpEqual, model.GetPrimaryKeyValue())); err != nil {
			a.marshalErrors(rw, 0, httputil.ErrInternalError())
			return
		}

		exists, err := db.Exists(params.Context, params.DB, existScope)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if !exists {
			a.marshalErrors(rw, http.StatusNotFound, httputil.ErrResourceNotFound())
			return
		}

		if removeRelationship {
			_, err = db.RemoveRelations(params.Context, params.DB, params.Scope, relation)
		} else {
			err = db.SetRelations(params.Context, params.DB, params.Scope, relation, payload.Data...)
		}
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		// Get and apply pre hook functions.
		for _, hook := range hooks.getPostHooksRelation(mStruct, query.UpdateRelationship, relation.Name()) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		var hasJsonapiMimeType bool
		for _, qv := range httputil.ParseAcceptHeader(req.Header) {
			if qv.Value == jsonapi.MimeType {
				hasJsonapiMimeType = true
				break
			}
		}

		if !hasJsonapiMimeType {
			log.Debug3f("[PATCH][%s][%s] No 'Accept' Header - returning HTTP Status: No Content - 204", mStruct.Collection(), s.ID)
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		link := codec.RelationshipLink
		if !a.MarshalLinks {
			link = codec.NoLink
		}
		resPayload := codec.Payload{
			ModelStruct: relation.Relationship().Struct(),
			Data:        payload.Data,
			FieldSets:   payload.FieldSets,
			MarshalLinks: &codec.LinkOptions{
				Type:          link,
				BaseURL:       a.BasePath,
				RootID:        id,
				Collection:    mStruct.Collection(),
				RelationField: relation.NeuronName(),
			},
			MarshalSingularFormat: relation.Kind() == mapping.KindRelationshipSingle,
		}
		a.marshalPayload(rw, &resPayload, http.StatusOK)
	}
}
