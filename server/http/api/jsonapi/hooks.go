package jsonapi

import (
	"context"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/db"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/server"
)

// QueryParams is the endpoint query parameters used in the whole endpoint.
type QueryParams struct {
	// Context is the context used in the endpoint r
	Context context.Context
	// DB is current interface used to get database connection. Overwritten to transaction would be used in other
	// database connections in given endpoint.
	DB db.DB
	// Scope is the query scope.
	Scope *query.Scope
	// Relations are the relations unmarshaled in the creation / update process.
	Relations     mapping.FieldSet
	Authenticator auth.Authenticator
	Authorizer    auth.Authorizer
}

// HookFunc is the function used as a route hook in the jsonapi handler.
type HookFunc func(p *QueryParams) error

var hooks *hooksStore

func init() {
	hooks = newHooksStore()
}

type hookKey struct {
	method   query.Method
	relation string
}

// AddPreHooks adds the pre method hook functions for general usage.
func (a *API) AddPreHooks(method query.Method, relationName string, hookFunctions ...HookFunc) {
	key := hookKey{method: method, relation: relationName}
	a.hooks.PreGeneralHooks[key] = append(hooks.PreGeneralHooks[key], hookFunctions...)
}

// AddPostHooks adds the post method hook functions for general usage.
func (a *API) AddPostHooks(method query.Method, relationName string, hookFunctions ...HookFunc) {
	key := hookKey{method: method, relation: relationName}
	a.hooks.PostGeneralHooks[key] = append(hooks.PostGeneralHooks[key], hookFunctions...)
}

// AddModelPreHooks adds the pre method hook functions for provided model.
func (a *API) AddModelPreHooks(model mapping.Model, method query.Method, hookFunctions ...HookFunc) error {
	return a.addModelPreHooks(model, method, "", hookFunctions)
}

// AddModelPreRelationHooks adds the hooks before database operation in the endpoint. The 'relation' is the golang structure
// field name that represents given relation.
func (a *API) AddModelPreRelationHooks(model mapping.Model, method query.Method, relation string, hookFunctions ...HookFunc) error {
	switch method {
	case query.InsertRelationship, query.GetRelated, query.GetRelationship, query.DeleteRelationship, query.UpdateRelationship:
	default:
		return errors.New(server.ClassEndpoint, "invalid query method for the related hook")
	}
	return a.addModelPreHooks(model, method, relation, hookFunctions)
}

func (a *API) addModelPreHooks(model mapping.Model, method query.Method, relationName string, hookFunctions []HookFunc) error {
	modelHooks := hooks.PreModelHooks[model.NeuronCollectionName()]
	key := hookKey{method: method, relation: relationName}
	if modelHooks == nil {
		modelHooks = map[hookKey][]HookFunc{}
	}
	modelHooks[key] = append(modelHooks[key], hookFunctions...)
	hooks.PreModelHooks[model.NeuronCollectionName()] = modelHooks
	return nil
}

// AddModelPostHooks adds the hooks post getting the database model function.
func (a *API) AddModelPostHooks(model mapping.Model, method query.Method, hookFunctions ...HookFunc) error {
	return a.addModelPostHooks(model, method, "", hookFunctions)
}

// AddModelPostRelationHooks adds the hooks after orm operation in the endpoint. The 'relation' is the golang structure
// field name that represents given relation.
func (a *API) AddModelPostRelationHooks(model mapping.Model, method query.Method, relation string, hookFunctions ...HookFunc) error {
	switch method {
	case query.InsertRelationship, query.GetRelated, query.GetRelationship, query.DeleteRelationship, query.UpdateRelationship:
	default:
		return errors.New(server.ClassEndpoint, "invalid query method for the related hook")
	}
	return a.addModelPostHooks(model, method, relation, hookFunctions)
}

func (a *API) addModelPostHooks(model mapping.Model, method query.Method, relationName string, hookFunctions []HookFunc) error {
	modelHooks := hooks.PostModelHooks[model.NeuronCollectionName()]
	if modelHooks == nil {
		modelHooks = map[hookKey][]HookFunc{}
	}
	key := hookKey{method: method, relation: relationName}
	modelHooks[key] = append(modelHooks[key], hookFunctions...)
	hooks.PostModelHooks[model.NeuronCollectionName()] = modelHooks
	return nil
}

type hooksStore struct {
	PreGeneralHooks  map[hookKey][]HookFunc
	PostGeneralHooks map[hookKey][]HookFunc
	PreModelHooks    map[string]map[hookKey][]HookFunc
	PostModelHooks   map[string]map[hookKey][]HookFunc
}

func newHooksStore() *hooksStore {
	return &hooksStore{
		PostGeneralHooks: map[hookKey][]HookFunc{},
		PreGeneralHooks:  map[hookKey][]HookFunc{},
		PreModelHooks:    map[string]map[hookKey][]HookFunc{},
		PostModelHooks:   map[string]map[hookKey][]HookFunc{},
	}
}

func (h *hooksStore) getPreHooks(model *mapping.ModelStruct, method query.Method) []HookFunc {
	key := hookKey{method: method}
	hooks := h.PreGeneralHooks[key]
	if modelHooks := h.PreModelHooks[model.Collection()]; modelHooks != nil {
		hooks = append(hooks, modelHooks[key]...)
	}
	return hooks
}

func (h *hooksStore) getPreHooksRelation(model *mapping.ModelStruct, method query.Method, relation string) []HookFunc {
	key := hookKey{method: method, relation: relation}
	hooks, ok := h.PreGeneralHooks[key]
	if !ok {
		hooks = h.PreGeneralHooks[hookKey{method: method}]
	}
	if modelHooks := h.PreModelHooks[model.Collection()]; modelHooks != nil {
		modelHooksSlice, ok := modelHooks[key]
		if !ok {
			modelHooksSlice = modelHooks[hookKey{method: method}]
		}
		hooks = append(hooks, modelHooksSlice...)
	}
	return hooks
}

func (h *hooksStore) getPostHooks(model *mapping.ModelStruct, method query.Method) []HookFunc {
	key := hookKey{method: method}
	hooks := h.PostGeneralHooks[key]
	if modelHooks := h.PostModelHooks[model.Collection()]; modelHooks != nil {
		hooks = append(hooks, modelHooks[key]...)
	}
	return hooks
}

func (h *hooksStore) getPostHooksRelation(model *mapping.ModelStruct, method query.Method, relation string) []HookFunc {
	key := hookKey{method: method, relation: relation}
	hooks, ok := h.PostGeneralHooks[key]
	if !ok {
		hooks = h.PostGeneralHooks[hookKey{method: method}]
	}
	if modelHooks := h.PostModelHooks[model.Collection()]; modelHooks != nil {
		modelHooksSlice, ok := modelHooks[key]
		if !ok {
			modelHooksSlice = modelHooks[hookKey{method: method}]
		}
		hooks = append(hooks, modelHooksSlice...)
	}
	return hooks
}
