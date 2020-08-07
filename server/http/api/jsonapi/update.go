package jsonapi

import (
	"net/http"

	"github.com/neuronlabs/neuron-extensions/codec/jsonapi"
	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
	"github.com/neuronlabs/neuron/server"
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
		pu := jsonapi.GetCodec(a.Controller).(codec.PayloadUnmarshaler)
		payload, err := pu.UnmarshalPayload(req.Body, codec.UnmarshalOptions{StrictUnmarshal: a.Options.StrictUnmarshal, ModelStruct: mStruct})
		if err != nil {
			log.Debugf("Unmarshal scope for: '%s' failed: %v", mStruct.Collection(), err)
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		switch len(payload.Data) {
		case 0:
			err := httputil.ErrInvalidInput()
			err.Detail = "no models found in the input"
			a.marshalErrors(rw, 0, err)
			return
		case 1:
		default:
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

		unmarshaledFieldset := payload.FieldSets[0]
		relations := mapping.FieldSet{}
		fields := mapping.FieldSet{}
		for _, field := range unmarshaledFieldset {
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
		payload.FieldSets[0] = fields
		for _, relation := range relations {
			payload.IncludedRelations = append(payload.IncludedRelations, &query.IncludedRelation{StructField: relation})
		}

		ctx := req.Context()
		db := a.DB
		if len(relations) > 0 {
			db, err = database.Begin(ctx, db, nil)
			if err != nil {
				log.Errorf("Can't begin transaction: %v", err)
				a.marshalErrors(rw, 0, httputil.ErrInternalError())
				return
			}
		}

		// Rollback the transaction if it exists and is not done yet.
		defer func() {
			tx, ok := db.(*database.Tx)
			if ok && !tx.Transaction.State.Done() {
				if err = tx.Rollback(); err != nil {
					log.Errorf("Rolling back transaction failed: %v", err)
				}
			}
		}()

		params := &server.Params{
			DB:            db,
			Ctx:           ctx,
			Authenticator: a.Authenticator,
			Authorizer:    a.Authorizer,
		}

		// Get and apply pre hook functions.
		modelHandler, hasModelHandler := a.handlers[mStruct]

		// Execute before update hook.
		if hasModelHandler {
			beforeUpdateHandler, ok := modelHandler.(server.BeforeUpdateHandler)
			if ok {
				if err := beforeUpdateHandler.HandleBeforeUpdate(params, payload); err != nil {
					a.marshalErrors(rw, 0, httputil.MapError(err)...)
					return
				}
			}
		}

		updateHandler, ok := modelHandler.(server.UpdateHandler)
		if !ok {
			// If no update handler is found execute default handler.
			updateHandler = a.defaultHandler
		}
		// Execute update handler.
		result, err := updateHandler.HandleUpdate(params.Ctx, *params, payload)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if hasModelHandler {
			afterHandler, ok := modelHandler.(server.AfterUpdateHandler)
			if ok {
				if err = afterHandler.HandleAfterUpdate(params, result); err != nil {
					a.marshalErrors(rw, 0, httputil.MapError(err)...)
					return
				}
			}
		}

		var hasJsonapiMimeType bool
		for _, qv := range httputil.ParseAcceptHeader(req.Header) {
			if qv.Value == jsonapi.MimeType {
				hasJsonapiMimeType = true
				break
			}
		}

		// TODO: get metadata from result payload.
		if !hasJsonapiMimeType {
			log.Debug3f("[PATCH][%s] No 'Accept' Header - returning HTTP Status: No Content - 204", mStruct.Collection())
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		// Prepare the scope for the api.GetHandler.
		getScope := query.NewScope(mStruct)
		getScope.FieldSets = []mapping.FieldSet{mStruct.Fields()}
		getScope.Filter(filter.New(mStruct.Primary(), filter.OpEqual, model.GetPrimaryKeyValue()))

		for _, relation := range mStruct.RelationFields() {
			if err = getScope.Include(relation, relation.Relationship().RelatedModelStruct().Primary()); err != nil {
				log.Errorf("Can't include relation field to the get scope: %v", err)
				a.marshalErrors(rw, 0, httputil.ErrInternalError())
				return
			}
		}

		// If a user had closed the transaction
		if tx, ok := params.DB.(*database.Tx); ok && tx.State().Done() {
			params.DB = a.DB
		}
		getResult, err := a.getHandleChain(params, getScope)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		linkType := codec.ResourceLink
		// but if the config doesn't allow that - set 'jsonapi.NoLink'
		if !a.Options.PayloadLinks {
			linkType = codec.NoLink
		}

		getResult.ModelStruct = mStruct
		getResult.FieldSets = []mapping.FieldSet{append(getScope.FieldSets[0], mStruct.RelationFields()...)}
		if getResult.MarshalLinks.Type == codec.NoLink {
			getResult.MarshalLinks = codec.LinkOptions{
				Type:       linkType,
				BaseURL:    a.Options.PathPrefix,
				RootID:     httputil.CtxMustGetID(ctx),
				Collection: mStruct.Collection(),
			}
		}
		getResult.MarshalSingularFormat = true
		a.marshalPayload(rw, getResult, http.StatusOK)
	}
}
