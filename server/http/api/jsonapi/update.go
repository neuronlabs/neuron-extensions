package jsonapi

import (
	"net/http"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/db"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"

	"github.com/neuronlabs/neuron-extensions/codec/jsonapi"
	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
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
		// unmarshal the input from the request body.
		pu := jsonapi.Codec().(codec.PayloadUnmarshaler)
		payload, err := pu.UnmarshalPayload(req.Body, codec.UnmarshalOptions{StrictUnmarshal: a.StrictUnmarshal, ModelStruct: mStruct})
		if err != nil {
			log.Debugf("Unmarshal scope for: '%s' failed: %v", mStruct.Collection(), err)
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if len(payload.Data) == 0 {
			err := httputil.ErrInvalidInput()
			err.Detail = "no models found in the input"
			a.marshalErrors(rw, 0, err)
			return
		}

		if len(payload.Data) > 1 {
			err := httputil.ErrInvalidInput()
			err.Detail = "bulk update is not implemented yet"
			a.marshalErrors(rw, 0, err)
			return
		}

		model := payload.Data[0]
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
		for _, field := range payload.FieldSets[0] {
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
		s := query.NewScope(mStruct)
		s.FieldSets = []mapping.FieldSet{fields}

		params := &QueryParams{
			Context:   req.Context(),
			DB:        a.DB,
			Scope:     s,
			Relations: relations,
		}

		if len(relations) > 0 {
			params.DB, err = db.Begin(params.Context, params.DB, nil)
			if err != nil {
				log.Errorf("Can't begin transaction: %v", err)
				a.marshalErrors(rw, 0, httputil.ErrInternalError())
				return
			}
		}

		// Rollback the transaction if it exists and is not done yet.
		defer func() {
			tx, ok := params.DB.(*db.Tx)
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
		if _, err = db.Update(params.Context, params.DB, params.Scope); err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if len(relations) > 0 {
			for _, relation := range relations {
				switch relation.Relationship().Kind() {
				case mapping.RelHasOne:
					// SetRelations first clear the relationship and then add it - it is not required here as a hasOne
					// only needs to add new relation to it's value.
					if err = db.AddRelations(params.Context, params.DB, s, relation); err != nil {
						a.marshalErrors(rw, 0, httputil.MapError(err)...)
						return
					}
				default:
					if err = db.SetRelations(params.Context, params.DB, s, relation); err != nil {
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
		getScope.FieldSets = []mapping.FieldSet{mStruct.Fields()}
		if getScope.Filter(filter.New(mStruct.Primary(), filter.OpEqual, model.GetPrimaryKeyValue())); err != nil {
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

		if _, err := db.Get(params.Context, a.DB, getScope); err != nil {
			log.Debugf("[PATCH][%s][%s] Getting resource after patching failed: %v", mStruct.Collection(), s.ID, err)
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		linkType := codec.ResourceLink
		// but if the config doesn't allow that - set 'jsonapi.NoLink'
		if !a.MarshalLinks {
			linkType = codec.NoLink
		}

		// TODO: add relations.
		resPayload := codec.Payload{
			ModelStruct:       mStruct,
			Data:              getScope.Models,
			FieldSets:         getScope.FieldSets,
			IncludedRelations: getScope.IncludedRelations,
			MarshalLinks: &codec.LinkOptions{
				Type:       linkType,
				BaseURL:    a.BasePath,
				RootID:     httputil.CtxMustGetID(params.Context),
				Collection: mStruct.Collection(),
			},
			MarshalSingularFormat: true,
		}
		a.marshalPayload(rw, &resPayload, http.StatusOK)
	}
}
