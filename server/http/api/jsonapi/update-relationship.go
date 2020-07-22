package jsonapi

import (
	"fmt"
	"net/http"

	"github.com/neuronlabs/neuron-plugins/codec/jsonapi"
	"github.com/neuronlabs/neuron-plugins/server/http/httputil"
	"github.com/neuronlabs/neuron-plugins/server/http/log"
	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/orm"
	"github.com/neuronlabs/neuron/query"
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
		relations, err := jsonapi.Codec().UnmarshalModels(req.Body, relation.Relationship().Struct(), a.jsonapiUnmarshalOptions())
		if err != nil {
			clErr, ok := err.(errors.ClassError)
			if !ok || clErr.Class() != codec.ClassNullDataInput {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
			removeRelationship = true
		}

		if len(relations) == 0 {
			removeRelationship = true
		}

		s := query.NewScope(mStruct, model)

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
		if err = existScope.Filter(query.NewFilterField(mStruct.Primary(), query.OpEqual, model.GetPrimaryKeyValue())); err != nil {
			a.marshalErrors(rw, 0, httputil.ErrInternalError())
			return
		}

		exists, err := orm.Exists(params.Context, params.DB, existScope)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if !exists {
			a.marshalErrors(rw, http.StatusNotFound, httputil.ErrResourceNotFound())
			return
		}

		if removeRelationship {
			_, err = orm.RemoveRelations(params.Context, params.DB, params.Scope, relation)
		} else {
			err = orm.SetRelations(params.Context, params.DB, params.Scope, relation, relations...)
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

		qm := jsonapi.Codec().(codec.QueryMarshaler)
		relatedScope := query.NewScope(relation.Relationship().Struct(), relations...)
		relatedScope.FieldSet = mapping.FieldSet{relation.Relationship().Struct().Primary()}

		link := codec.RelationshipLink
		if !a.MarshalLinks {
			link = codec.NoLink
		}

		options := &codec.MarshalOptions{
			Link: codec.LinkOptions{
				Type:         link,
				BaseURL:      a.BasePath,
				RootID:       id,
				Collection:   mStruct.Collection(),
				RelatedField: relation.NeuronName(),
			},
		}
		qm, ok := jsonapi.Codec().(codec.QueryMarshaler)
		if !ok {
			log.Errorf("jsonapi codec is not a QueryMarshaler")
			a.marshalErrors(rw, 500, httputil.ErrInternalError())
			return
		}
		if err = qm.MarshalQuery(rw, relatedScope, options); err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}
	}
}
