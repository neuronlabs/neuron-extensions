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

// HandleInsert handles json:api post endpoint for the 'model'. Panics if the model is not mapped for given API controller.
func (a *API) HandleInsert(model mapping.Model) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		a.handleInsert(a.Controller.MustModelStruct(model))(rw, req)
	}
}

func (a *API) handleInsert(model *mapping.ModelStruct) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// unmarshal the input from the request body.
		qu := jsonapi.Codec().(codec.QueryUnmarshaler)
		q, err := qu.UnmarshalQuery(req.Body, model, &codec.UnmarshalOptions{StrictUnmarshal: a.StrictMarshal})
		if err != nil {
			log.Debugf("Unmarshal scope for: '%s' failed: %v", model.Collection(), err)
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if len(q.Models) == 0 {
			err := httputil.ErrInvalidInput()
			err.Detail = "nothing to insert"
			a.marshalErrors(rw, 0, err)
			return
		}

		// Divide fieldset into fields and relations.
		relations := mapping.FieldSet{}
		fields := mapping.FieldSet{}
		var isPrimary bool
		for _, field := range q.FieldSet {
			switch field.Kind() {
			case mapping.KindRelationshipMultiple, mapping.KindRelationshipSingle:
				// If the relationship is of BelongsTo kind - set its relationship primary key value into given model's foreign key.
				if field.Relationship().Kind() == mapping.RelBelongsTo {
					relationer, ok := q.Models[0].(mapping.SingleRelationer)
					if !ok {
						log.Errorf("Model: '%s' doesn't implement mapping.SingleRelationer interface", model.Collection())
						a.marshalErrors(rw, 500, httputil.ErrInternalError())
						return
					}
					relation, err := relationer.GetRelationModel(field)
					if err != nil {
						a.marshalErrors(rw, 0, httputil.MapError(err)...)
						return
					}
					fielder, ok := q.Models[0].(mapping.Fielder)
					if !ok {
						log.Errorf("Model: '%s' doesn't implement mapping.SingleRelationer interface", model.Collection())
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
			case mapping.KindPrimary:
				isPrimary = true
			}
			fields = append(fields, field)
		}

		// Check if a model is allowed to set it's primary key.
		if isPrimary && !model.AllowClientID() {
			log.Debug2f("Creating: '%s' with client-generated ID is forbidden", model.Collection())
			err := httputil.ErrInvalidJSONFieldValue()
			err.Detail = "Client-Generated ID is not allowed for this model."
			err.Status = "403"
			a.marshalErrors(rw, http.StatusForbidden, err)
			return
		}
		fields.Sort()

		q.FieldSet = fields

		// Create query parameters.
		params := &QueryParams{
			Context:   req.Context(),
			DB:        a.DB,
			Scope:     q,
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
		for _, hook := range hooks.getPreHooks(model, query.Insert) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		// Insert into database.
		if err = orm.Insert(params.Context, params.DB, params.Scope); err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if len(relations) > 0 {
			for _, relation := range relations {
				switch relation.Relationship().Kind() {
				case mapping.RelHasOne:
					// SetRelations first clear the relationship and then add it - it is not required here as a hasOne
					// only needs to add new relation to it's value.
					if err = orm.AddRelations(params.Context, params.DB, q, relation); err != nil {
						a.marshalErrors(rw, 0, httputil.MapError(err)...)
						return
					}
				default:
					if err = orm.SetRelations(params.Context, params.DB, q, relation); err != nil {
						a.marshalErrors(rw, 0, httputil.MapError(err)...)
						return
					}
				}
			}
		}

		// Get and apply pre hook functions.
		for _, hook := range hooks.getPostHooks(model, query.Insert) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}
		// if the primary was provided in the input and if the config doesn't allow to return
		// created value with given client-id - return simple status NoContent
		if isPrimary && a.NoContentOnCreate {
			// if the primary was provided
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		// get the primary field value so that it could be used for the jsonapi marshal process.
		stringID, err := q.Models[0].GetPrimaryKeyStringValue()
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		// By default marshal resource links
		linkType := codec.ResourceLink
		// but if the config doesn't allow that - set 'jsonapi.NoLink'
		if !a.MarshalLinks {
			linkType = codec.NoLink
		}
		// prepare the options to marshal jsonapi scope.
		options := &codec.MarshalOptions{
			Link: codec.LinkOptions{
				Type:       linkType,
				BaseURL:    a.getBasePath(a.BasePath),
				Collection: model.Collection(),
				RootID:     stringID,
			},
		}
		a.marshalScope(params.Scope, rw, http.StatusCreated, options)
	}
}
