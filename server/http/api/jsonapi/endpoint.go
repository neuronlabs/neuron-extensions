package jsonapi

import (
	"net/http"

	"github.com/neuronlabs/neuron-plugins/server/http/middleware"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/server"
)

// Endpoint is the structure used for creation of the
type Endpoint struct {
	Method query.Method
	// Relation is the relation field used for the 'GetRelated', 'GetRelationship', 'UpdateRelationship', 'DeleteRelationship' methods.
	Relation string
	// Middlewares are the middlewares used on that endpoint
	Middlewares middleware.Chain
	Handler     http.HandlerFunc
}

// RegisterEndpoint registers provided endpoint within given API.
func (a *API) RegisterEndpoint(model mapping.Model, endpoint *Endpoint) error {
	mStruct, err := a.Controller.ModelStruct(model)
	if err != nil {
		return err
	}
	switch endpoint.Method {
	case query.InsertRelationship, query.UpdateRelationship, query.DeleteRelationship:
		if endpoint.Relation == "" {
			return errors.NewDetf(server.ClassInternal, "jsonapi endpoint method: %v, for model: '%s' has empty relationship field: '%s'", endpoint.Method, mStruct)
		}
	case query.InvalidMethod:
		return errors.New(server.ClassInternal, "provided invalid query method for the endpoint")
	}

	endpoints, ok := a.modelEndpoints[model.NeuronCollectionName()]
	if ok {
		for _, existingEndpoint := range endpoints {
			if existingEndpoint.Method == endpoint.Method {
				switch endpoint.Method {
				case query.InsertRelationship, query.UpdateRelationship, query.DeleteRelationship:
					if endpoint.Relation == existingEndpoint.Relation {
						return errors.NewDetf(server.ClassInternal, "jsonapi endpoint method: %v, already registered for model: '%s' with field: '%s'", endpoint.Method, mStruct, endpoint.Relation)
					}
				default:
					return errors.NewDetf(server.ClassInternal, "jsonapi endpoint method: %v, already registered for model: '%s'", endpoint.Method, mStruct)
				}
			}
		}
	}
	endpoints = append(endpoints, endpoint)
	return nil
}

// RegisterModelEndpoints registers endpoints for given model.
func (a *API) RegisterModelEndpoints(model mapping.Model, endpoints ...*Endpoint) error {
	for _, endpoint := range endpoints {
		if err := a.RegisterEndpoint(model, endpoint); err != nil {
			return err
		}
	}
	return nil
}

// RegisterModelDefaultEndpoints sets default endpoints for 'model'. If some custom endpoints were set for the model,
// 'API' would use it's custom endpoint.
func (a *API) RegisterModelDefaultEndpoints(model mapping.Model) {
	a.defaultEndpoints[model.NeuronCollectionName()] = struct{}{}
}

var (
	basicMethods   = [...]query.Method{query.Insert, query.Get, query.List, query.Update, query.Delete}
	relatedMethods = [...]query.Method{query.InsertRelationship, query.GetRelationship, query.GetRelated, query.DeleteRelationship, query.UpdateRelationship}
)

func (a *API) setModelsDefaultEndpoints(mStruct *mapping.ModelStruct) {
	type endpoint struct {
		method   query.Method
		relation string
	}
	endpointsMap := map[endpoint]struct{}{}

	collectionName := mStruct.Collection()
	modelEndpoints := a.modelEndpoints[collectionName]
	for _, e := range a.modelEndpoints[collectionName] {
		endpointsMap[endpoint{method: e.Method, relation: e.Relation}] = struct{}{}
	}

	// Handle basic endopoints.
	for _, method := range basicMethods {
		if _, ok := endpointsMap[endpoint{method: method}]; !ok {
			modelEndpoints = append(modelEndpoints, &Endpoint{Method: method})
		}
	}
	// Handle relation endpoints.
	for _, relation := range mStruct.RelationFields() {
		for _, method := range relatedMethods {
			_, isNeuron := endpointsMap[endpoint{method: method, relation: relation.NeuronName()}]
			_, isNormal := endpointsMap[endpoint{method: method, relation: relation.Name()}]
			if !isNeuron && !isNormal {
				modelEndpoints = append(modelEndpoints, &Endpoint{Method: method, Relation: relation.NeuronName()})
			}
		}
	}
	a.modelEndpoints[collectionName] = modelEndpoints
}

func (a *API) setEndpointHandlerFunc(mStruct *mapping.ModelStruct, endpoint *Endpoint) error {
	switch endpoint.Method {
	case query.Insert:
		endpoint.Handler = a.handleInsert(mStruct)
		endpoint.Middlewares = append(middleware.Chain{MidContentType}, endpoint.Middlewares...)
	case query.InsertRelationship:
		relation, err := a.getRelation(endpoint, mStruct)
		if err != nil {
			return err
		}
		endpoint.Handler = a.handleInsertRelationship(mStruct, relation)
		endpoint.Middlewares = append(middleware.Chain{MidContentType, middleware.StoreIDFromParams("id")}, endpoint.Middlewares...)
	case query.Delete:
		endpoint.Handler = a.handleDelete(mStruct)
		endpoint.Middlewares = append(middleware.Chain{middleware.StoreIDFromParams("id")})
	case query.DeleteRelationship:
		relation, err := a.getRelation(endpoint, mStruct)
		if err != nil {
			return err
		}
		endpoint.Handler = a.handleDeleteRelationship(mStruct, relation)
		endpoint.Middlewares = append(middleware.Chain{MidContentType, middleware.StoreIDFromParams("id")}, endpoint.Middlewares...)
	case query.Get:
		endpoint.Handler = a.handleGet(mStruct)
		endpoint.Middlewares = append(middleware.Chain{MidAccept, middleware.StoreIDFromParams("id")}, endpoint.Middlewares...)
	case query.GetRelated:
		relation, err := a.getRelation(endpoint, mStruct)
		if err != nil {
			return err
		}
		endpoint.Handler = a.handleGetRelated(mStruct, relation)
		endpoint.Middlewares = append(middleware.Chain{MidAccept, middleware.StoreIDFromParams("id")}, endpoint.Middlewares...)
	case query.GetRelationship:
		relation, err := a.getRelation(endpoint, mStruct)
		if err != nil {
			return err
		}
		endpoint.Handler = a.handleGetRelationship(mStruct, relation)
		endpoint.Middlewares = append(middleware.Chain{MidAccept, middleware.StoreIDFromParams("id")}, endpoint.Middlewares...)
	case query.List:
		endpoint.Handler = a.handleList(mStruct)
		endpoint.Middlewares = append(middleware.Chain{MidAccept}, endpoint.Middlewares...)
	case query.Update:
		endpoint.Handler = a.handleUpdate(mStruct)
		endpoint.Middlewares = append(middleware.Chain{MidContentType, middleware.StoreIDFromParams("id")}, endpoint.Middlewares...)
	case query.UpdateRelationship:
		relation, err := a.getRelation(endpoint, mStruct)
		if err != nil {
			return err
		}
		endpoint.Handler = a.handleUpdateRelationship(mStruct, relation)
		endpoint.Middlewares = append(middleware.Chain{MidContentType, middleware.StoreIDFromParams("id")}, endpoint.Middlewares...)
	}
	return nil
}
