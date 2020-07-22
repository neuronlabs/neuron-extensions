package jsonapi

import (
	"net/http"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/orm"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-plugins/codec/jsonapi"
	"github.com/neuronlabs/neuron-plugins/server/http/httputil"
	"github.com/neuronlabs/neuron-plugins/server/http/log"
)

// HandleUpdate handles json:api list endpoint for the 'model'. Panics if the model is not mapped for given API controller.
func (a *API) HandleUpdate(model mapping.Model) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		a.handleUpdate(a.Controller.MustModelStruct(model))(rw, req)
	}
}

func (a *API) handleUpdate(mStruct *mapping.ModelStruct) http.HandlerFunc {
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

		qu := jsonapi.Codec().(codec.QueryUnmarshaler)
		s, err := qu.UnmarshalQuery(req.Body, mStruct, a.jsonapiUnmarshalOptions())
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if len(s.Models) == 0 {
			err := httputil.ErrInvalidInput()
			err.Detail = "no models found in the input"
			a.marshalErrors(rw, 0, err)
			return
		}

		model := s.Models[0]
		if model.IsPrimaryKeyZero() {
			err = model.SetPrimaryKeyStringValue(id)
		} else {
			unmarshaledID, err := model.GetPrimaryKeyStringValue()
			if err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
			if unmarshaledID != id {
				err := httputil.ErrInvalidInput()
				err.Detail = "provided input model 'id' differs from the one in the URI"
				log.Debug2f("[PATCH][%s] %s", mStruct.Collection(), err.Detail)
				a.marshalErrors(rw, 0, err)
				return
			}
		}

		relations := mapping.FieldSet{}
		fields := mapping.FieldSet{}
		for _, field := range s.FieldSet {
			switch field.Kind() {
			case mapping.KindRelationshipMultiple, mapping.KindRelationshipSingle:
				// If the relationship is of BelongsTo kind - set its relationship primary key value into given model's foreign key.
				if field.Relationship().Kind() == mapping.RelBelongsTo {
					relationer, ok := model.(mapping.SingleRelationer)
					if !ok {
						log.Errorf("Model: '%s' doesn't implement mapping.SingleRelationer interface", mStruct.Collection())
						a.marshalErrors(rw, 500, httputil.ErrInternalError())
						return
					}
					relation, err := relationer.GetRelationModel(field)
					if err != nil {
						a.marshalErrors(rw, 0, httputil.MapError(err)...)
						return
					}
					fielder, ok := model.(mapping.Fielder)
					if !ok {
						log.Errorf("Model: '%s' doesn't implement mapping.SingleRelationer interface", mStruct.Collection())
						a.marshalErrors(rw, 500, httputil.ErrInternalError())
						return
					}
					if err = fielder.SetFieldValue(field.Relationship().ForeignKey(), relation.GetPrimaryKeyValue()); err != nil {
						a.marshalErrors(rw, 0, httputil.MapError(err)...)
						return
					}
					fields = append(fields, field.Relationship().ForeignKey())
					continue
				}
				// All the other foreign relations should be post insert.
				relations = append(relations, field)
				continue
			}
			fields = append(fields, field)
		}

		fields.Sort()
		s.FieldSet = fields

		params := &QueryParams{
			Context:   req.Context(),
			DB:        a.DB,
			Scope:     s,
			Relations: relations,
		}

		if len(relations) > 0 {
			params.DB, err = orm.Begin(params.Context, params.DB, nil)
			if err != nil {
				log.Errorf("Can't begin transaction: %v", err)
				a.marshalErrors(rw, 0, httputil.ErrInternalError())
				return
			}
		}

		// Rollback the transaction if it exists and is not done yet.
		defer func() {
			tx, ok := params.DB.(*orm.Tx)
			if ok && !tx.Transaction.State.Done() {
				if err = tx.Rollback(); err != nil {
					log.Errorf("Rolling back transaction failed: %v", err)
				}
			}
		}()

		// Get and apply pre hook functions.
		for _, hook := range hooks.getPreHooks(mStruct, query.Update) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		// Insert into database.
		if _, err = orm.Update(params.Context, params.DB, params.Scope); err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if len(relations) > 0 {
			for _, relation := range relations {
				switch relation.Relationship().Kind() {
				case mapping.RelHasOne:
					// SetRelations first clear the relationship and then add it - it is not required here as a hasOne
					// only needs to add new relation to it's value.
					if err = orm.AddRelations(params.Context, params.DB, s, relation); err != nil {
						a.marshalErrors(rw, 0, httputil.MapError(err)...)
						return
					}
				default:
					if err = orm.SetRelations(params.Context, params.DB, s, relation); err != nil {
						a.marshalErrors(rw, 0, httputil.MapError(err)...)
						return
					}
				}
			}
		}

		// Get and apply pre hook functions.
		for _, hook := range hooks.getPostHooks(mStruct, query.Update) {
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

		getScope := query.NewScope(mStruct)
		getScope.FieldSet = mStruct.Fields()
		if err = getScope.Filter(query.NewFilterField(mStruct.Primary(), query.OpEqual, model.GetPrimaryKeyValue())); err != nil {
			log.Errorf("[PATCH][SCOPE][%s] Adding param primary filter to return content scope failed: %v", err)
			a.marshalErrors(rw, 0, httputil.ErrInternalError())
			return
		}

		for _, relation := range mStruct.RelationFields() {
			if err = getScope.Include(relation, relation.Relationship().Struct().Primary()); err != nil {
				log.Errorf("Can't include relation field to the get scope: %v", err)
				a.marshalErrors(rw, 0, httputil.ErrInternalError())
				return
			}
		}

		if _, err := orm.Get(params.Context, a.DB, getScope); err != nil {
			log.Debugf("[PATCH][%s][%s] Getting resource after patching failed: %v", mStruct.Collection(), s.ID, err)
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		options := &codec.MarshalOptions{Link: codec.LinkOptions{
			Type:       codec.ResourceLink,
			BaseURL:    a.BasePath,
			RootID:     id,
			Collection: mStruct.Collection(),
		}}
		a.marshalScope(getScope, rw, http.StatusOK, options)
	}
}
