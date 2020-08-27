package httputil

import (
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
	"github.com/neuronlabs/neuron/server"
)

// ErrFunc is the function used to create new Error instance.
type ErrFunc func() *codec.Error

// DefaultErrorMapper is the default error classification mapper.
var DefaultErrorMapper = &ErrorMapper{
	errors.ErrInternal: ErrInternalError,
	// Codec:
	codec.ErrMarshal:             ErrInternalError,
	codec.ErrUnmarshalDocument:   ErrInvalidJSONDocument,
	codec.ErrCodec:               ErrInvalidInput,
	codec.ErrUnmarshal:           ErrInvalidInput,
	codec.ErrUnmarshalFieldValue: ErrInvalidJSONFieldValue,
	codec.ErrUnmarshalFieldName:  ErrInvalidJSONFieldName,
	query.ErrViolation:           ErrInvalidJSONFieldValue,
	// Filter
	filter.ErrFilter:           ErrInvalidQueryParameter,
	filter.ErrFilterCollection: ErrInvalidResourceName,
	// Query:
	query.ErrFieldValue:         ErrInvalidJSONFieldValue,
	query.ErrInvalidParameter:   ErrInvalidQueryParameter,
	query.ErrInvalidSort:        ErrInvalidQueryParameter,
	query.ErrInvalidFieldSet:    ErrInvalidQueryParameter,
	query.ErrNoFieldsInFieldSet: ErrInvalidQueryParameter,
	query.ErrViolationUnique:    ErrResourceAlreadyExists,
	query.ErrViolationCheck:     ErrInvalidJSONFieldValue,
	query.ErrViolationNotNull:   ErrInvalidInput,
	// Mapping:
	mapping.ErrModel:                ErrInternalError,
	mapping.ErrRelation:             ErrInternalError,
	mapping.ErrFieldValue:           ErrInvalidJSONFieldValue,
	mapping.ErrFieldNotParser:       ErrInvalidQueryParameter,
	mapping.ErrFieldNotNullable:     ErrForbiddenOperation,
	mapping.ErrModelNotFound:        ErrResourceNotFound,
	mapping.ErrInvalidRelationField: ErrInvalidQueryParameter,
	// Server:
	server.ErrHeaderNotAcceptable:   ErrNotAcceptable,
	server.ErrUnsupportedHeader:     ErrUnsupportedHeader,
	server.ErrMissingRequiredHeader: ErrMissingRequiredHeader,
	server.ErrHeaderValue:           ErrInvalidHeaderValue,
	// Auth:
	auth.ErrAccountNotFound:      ErrInvalidAuthenticationInfo,
	auth.ErrAccountAlreadyExists: ErrAccountAlreadyExists,
	auth.ErrInvalidUsername:      ErrInvalidAuthenticationInfo,
	auth.ErrInvalidSecret:        ErrInvalidAuthenticationInfo,
	auth.ErrInvalidPassword:      ErrInvalidAuthenticationInfo,
	auth.ErrTokenRevoked:         ErrInvalidAuthenticationInfo,
	auth.ErrTokenNotValidYet:     ErrInvalidAuthenticationInfo,
	auth.ErrTokenExpired:         ErrInvalidAuthenticationInfo,
	auth.ErrForbidden:            ErrForbiddenAuthorize,
	auth.ErrInvalidRole:          ErrForbiddenAuthorize,
	auth.ErrAuthorizationHeader:  ErrInvalidAuthorizationHeader,
}

// MapError maps the 'err' input error into slice of 'Error'.
// The function uses DefaultErrorMapper for error mapping.
// The logic is the same as for DefaultErrorMapper.Errors method.
func MapError(err error) []*codec.Error {
	return DefaultErrorMapper.errors(err)
}

// ErrorMapper is the neuron errors classification mapper.
// It creates the 'Error' from the provided error.
type ErrorMapper map[error]ErrFunc

// Errors gets the slice of 'Error' from the provided 'err' error.
// The mapping is based on the 'most specific classification first' method.
// If the error is 'errors.ClassError' the function gets it's class.
// The function checks classification occurrence based on the priority:
//	- Class
//	- Minor
//	- Major
// If no mapping is provided for given classification - an internal error is returned.
func (c *ErrorMapper) Errors(err error) []*codec.Error {
	return c.errors(err)
}

func (c *ErrorMapper) errors(err error) []*codec.Error {
	switch et := err.(type) {
	case *codec.Error:
		return []*codec.Error{et}
	case codec.MultiError:
		return et
	case errors.MultiError:
		var errs []*codec.Error
		for _, single := range et {
			errs = append(errs, c.mapSingleError(single))
		}
		return errs
	default:
		log.Debugf("Unknown error: %+v", err)
	}
	return []*codec.Error{ErrInternalError()}
}

func (c *ErrorMapper) mapSingleError(e error) *codec.Error {
	// check if the class is stored in the mapper
	err := e
	for {
		creator, ok := (*c)[e]
		if ok {
			cErr := creator()
			detailed, ok := err.(*errors.DetailedError)
			if ok {
				cErr.Detail = detailed.Details
				cErr.ID = detailed.ID.String()
			}
			return cErr
		}
		if e = errors.Unwrap(e); e == nil {
			return ErrInternalError()
		}
	}
}
