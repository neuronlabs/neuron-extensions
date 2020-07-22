package jsonapi

import (
	"net/http"
	"net/url"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/orm"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-plugins/codec/jsonapi"
	"github.com/neuronlabs/neuron-plugins/server/http/httputil"
	"github.com/neuronlabs/neuron-plugins/server/http/log"
)

// HandleList handles json:api list endpoint for the 'model'. Panics if the model is not mapped for given API controller.
func (a *API) HandleList(model mapping.Model) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		a.handleList(a.Controller.MustModelStruct(model))(rw, req)
	}
}

func (a *API) handleList(model *mapping.ModelStruct) http.HandlerFunc {
	var defaultPagination *query.Pagination
	if a.DefaultPageSize > 0 {
		defaultPagination = &query.Pagination{
			Limit:  int64(a.DefaultPageSize),
			Offset: 0,
		}
		log.Debug2f("Default pagination at 'GET /%s' is: %v", model.Collection(), defaultPagination.String())
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		s, err := a.createListScope(model, req)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		if defaultPagination != nil && s.Pagination == nil {
			s.Pagination = &(*defaultPagination)
		}

		if log.Level() >= log.LevelDebug3 {
			log.Debug3f("[LIST] %s", s.String())
		}

		// Create query params.
		params := &QueryParams{
			Context: req.Context(),
			Scope:   s,
			DB:      a.DB,
		}

		// Get and apply pre hook functions.
		for _, hook := range hooks.getPreHooks(model, query.List) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		if _, err := orm.Find(params.Context, params.DB, params.Scope); err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		// Get and apply post hook functions.
		for _, hook := range hooks.getPostHooks(model, query.List) {
			if err = hook(params); err != nil {
				a.marshalErrors(rw, 0, httputil.MapError(err)...)
				return
			}
		}

		linkType := codec.ResourceLink
		if !a.MarshalLinks {
			linkType = codec.NoLink
		}
		options := &codec.MarshalOptions{Link: codec.LinkOptions{
			Type:       linkType,
			BaseURL:    a.BasePath,
			Collection: model.Collection(),
		}}

		// if there were a query no set link type to 'NoLink'
		if v, ok := s.StoreGet(jsonapi.StoreKeyMarshalLinks); ok && !v.(bool) {
			options.Link.Type = codec.NoLink
		}

		// if there is no pagination then the pagination doesn't need to be created.
		// marshal the results if there were no pagination set
		if s.Pagination == nil || len(s.Models) == 0 {
			a.marshalScope(s, rw, http.StatusOK, options)
			return
		}

		// prepare new count scope - and build query parameters for the pagination
		// page[limit] page[offset] page[number] page[size]
		countScope := s.Copy()
		total, err := orm.Count(params.Context, a.DB, countScope)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}

		temp := a.queryWithoutPagination(req)

		// extract query values from the req.URL
		// prepare the pagination links for the options
		s.Pagination.FormatQuery(temp)

		paginationLinks := &codec.PaginationLinks{Total: total}
		options.Link.PaginationLinks = paginationLinks
		options.Link.PaginationLinks.Self = temp.Encode()

		next, err := s.Pagination.Next(total)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}
		temp = a.queryWithoutPagination(req)

		if next != s.Pagination {
			next.FormatQuery(temp)
			paginationLinks.Next = temp.Encode()
			temp = url.Values{}
		}

		prev, err := s.Pagination.Previous()
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}
		if prev != s.Pagination {
			prev.FormatQuery(temp)
			paginationLinks.Prev = temp.Encode()
			temp = a.queryWithoutPagination(req)
		}

		last, err := s.Pagination.Last(total)
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}
		last.FormatQuery(temp)

		paginationLinks.Last = temp.Encode()

		temp = a.queryWithoutPagination(req)
		first, err := s.Pagination.First()
		if err != nil {
			a.marshalErrors(rw, 0, httputil.MapError(err)...)
			return
		}
		first.FormatQuery(temp)
		paginationLinks.First = temp.Encode()
		a.marshalScope(s, rw, http.StatusOK, options)
	}
}

func (a *API) queryWithoutPagination(req *http.Request) url.Values {
	temp := url.Values{}

	for k, v := range req.URL.Query() {
		switch k {
		case query.ParamPageLimit, jsonapi.ParamPageNumber, query.ParamPageOffset, jsonapi.ParamPageSize:
		default:
			temp[k] = v
		}
	}
	return temp
}
