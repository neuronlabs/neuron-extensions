package jsonapi

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/neuronlabs/neuron-extensions/codec/jsonapi"
	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/server"
)

// HandleList handles json:api list endpoint for the 'model'. Panics if the model is not mapped for given API controller.
func (a *API) HandleList(model mapping.Model) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		a.handleList(a.Controller.MustModelStruct(model))(rw, req)
	}
}

func (a *API) handleList(mStruct *mapping.ModelStruct) http.HandlerFunc {
	var defaultPagination *query.Pagination
	if a.Options.DefaultPageSize > 0 {
		defaultPagination = &query.Pagination{
			Limit:  int64(a.Options.DefaultPageSize),
			Offset: 0,
		}
		log.Debug2f("Default pagination at 'GET /%s' is: %v", mStruct.Collection(), defaultPagination.String())
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		s, err := a.createListScope(mStruct, req)
		if err != nil {
			log.Debugf("[LIST][%s] parsing request query failed: %v", mStruct, err)
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if defaultPagination != nil && s.Pagination == nil {
			s.Pagination = &(*defaultPagination)
		}

		// queryIncludes are the included fields from the url query.
		queryIncludes := s.IncludedRelations
		var queryFieldSet mapping.FieldSet
		var fields mapping.FieldSet
		if len(s.FieldSets) == 0 {
			fields = append(s.ModelStruct.Attributes(), s.ModelStruct.RelationFields()...)
			queryFieldSet = fields
		} else {
			fields = s.FieldSets[0]
			queryFieldSet = s.FieldSets[0]
		}
		// json:api fieldset is a combination of fields + relations.
		// The same situation is with includes.
		neuronFields, neuronIncludes := parseFieldSetAndIncludes(mStruct, fields, queryIncludes)
		s.FieldSets = []mapping.FieldSet{neuronFields}
		s.IncludedRelations = neuronIncludes

		// Create query params.
		params := &server.Params{
			Ctx:           req.Context(),
			DB:            a.DB,
			Authenticator: a.Authenticator,
			Authorizer:    a.Authorizer,
		}

		// Get and apply pre hook functions.
		result, err := a.listHandleChain(params, s)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		linkType := codec.ResourceLink
		if !a.Options.PayloadLinks {
			linkType = codec.NoLink
		}

		// if there were a query no set link type to 'NoLink'
		if v, ok := s.StoreGet(jsonapi.StoreKeyMarshalLinks); ok {
			if v.(bool) {
				linkType = codec.ResourceLink
			} else {
				linkType = codec.NoLink
			}
		}

		result.ModelStruct = mStruct
		result.IncludedRelations = queryIncludes
		result.FieldSets = []mapping.FieldSet{queryFieldSet}
		if result.MarshalLinks.Type == codec.NoLink {
			result.MarshalLinks = codec.LinkOptions{
				Type:       linkType,
				BaseURL:    a.Options.PathPrefix,
				Collection: mStruct.Collection(),
			}
		}

		// if there is no pagination then the pagination doesn't need to be created.
		// marshal the results if there were no pagination set
		if s.Pagination == nil || len(s.Models) == 0 {
			result.PaginationLinks = &codec.PaginationLinks{}
			sb := strings.Builder{}
			sb.WriteString(a.basePath())
			sb.WriteRune('/')
			sb.WriteString(mStruct.Collection())
			if q := req.URL.Query(); len(q) > 0 {
				sb.WriteRune('?')
				sb.WriteString(q.Encode())
			}
			result.PaginationLinks.Self = sb.String()
			a.marshalPayload(rw, result, http.StatusOK)
			return
		}

		// prepare new count scope - and build query parameters for the pagination.
		// page[limit] page[offset] page[number] page[size]
		countScope := s.Copy()
		total, err := database.Count(req.Context(), a.DB, countScope)
		if err != nil {
			log.Debugf("[LIST][%s] Getting total values for given query failed: %v", mStruct, err)
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		temp, pageBased := a.queryWithoutPagination(req)

		// extract query values from the req.URL
		// prepare the pagination links for the options
		jsonapi.FormatPagination(s.Pagination, temp, pageBased)

		paginationLinks := &codec.PaginationLinks{Total: total}
		sb := strings.Builder{}
		sb.WriteString(a.basePath())
		sb.WriteRune('/')
		sb.WriteString(mStruct.Collection())
		sb.WriteRune('?')
		sb.WriteString(temp.Encode())
		paginationLinks.Self = sb.String()
		sb.Reset()

		next, err := s.Pagination.Next(total)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}
		temp, _ = a.queryWithoutPagination(req)

		if next != s.Pagination {
			jsonapi.FormatPagination(next, temp, pageBased)
			sb.WriteString(a.basePath())
			sb.WriteRune('/')
			sb.WriteString(mStruct.Collection())
			sb.WriteRune('?')
			sb.WriteString(temp.Encode())
			paginationLinks.Next = sb.String()
			sb.Reset()
			temp, _ = a.queryWithoutPagination(req)
		}

		prev, err := s.Pagination.Previous()
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}
		if prev != s.Pagination {
			jsonapi.FormatPagination(prev, temp, pageBased)
			sb.WriteString(a.basePath())
			sb.WriteRune('/')
			sb.WriteString(mStruct.Collection())
			sb.WriteRune('?')
			sb.WriteString(temp.Encode())
			paginationLinks.Prev = sb.String()
			sb.Reset()
			temp, _ = a.queryWithoutPagination(req)
		}

		last, err := s.Pagination.Last(total)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}
		jsonapi.FormatPagination(last, temp, pageBased)
		sb.WriteString(a.basePath())
		sb.WriteRune('/')
		sb.WriteString(mStruct.Collection())
		sb.WriteRune('?')
		sb.WriteString(temp.Encode())
		paginationLinks.Last = sb.String()
		sb.Reset()

		temp, _ = a.queryWithoutPagination(req)
		first, err := s.Pagination.First()
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}
		jsonapi.FormatPagination(first, temp, pageBased)
		sb.WriteString(a.basePath())
		sb.WriteRune('/')
		sb.WriteString(mStruct.Collection())
		sb.WriteRune('?')
		sb.WriteString(temp.Encode())
		paginationLinks.First = sb.String()

		result.PaginationLinks = paginationLinks
		a.marshalPayload(rw, result, http.StatusOK)
	}
}

func (a *API) queryWithoutPagination(req *http.Request) (url.Values, bool) {
	temp := url.Values{}
	var pageBased bool
	for k, v := range req.URL.Query() {
		switch k {
		case query.ParamPageLimit, query.ParamPageOffset:
		case jsonapi.ParamPageNumber, jsonapi.ParamPageSize:
			pageBased = true
		default:
			temp[k] = v
		}
	}
	return temp, pageBased
}

func (a *API) listHandleChain(params *server.Params, q *query.Scope) (*codec.Payload, error) {
	modelHandler, hasModelHandler := a.handlers[q.ModelStruct]
	if hasModelHandler {
		beforeHandler, ok := modelHandler.(server.BeforeListHandler)
		if ok {
			if err := beforeHandler.HandleBeforeList(params, q); err != nil {
				return nil, err
			}
		}
	}

	getHandler, ok := modelHandler.(server.ListHandler)
	if !ok {
		getHandler = a.defaultHandler
	}
	result, err := getHandler.HandleList(params.Ctx, *params, q)
	if err != nil {
		return nil, err
	}

	if hasModelHandler {
		afterHandler, ok := modelHandler.(server.AfterListHandler)
		if ok {
			if err := afterHandler.HandleAfterList(params, result); err != nil {
				return nil, err
			}
		}
	}
	return result, err
}
