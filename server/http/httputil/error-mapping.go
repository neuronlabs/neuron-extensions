package httputil

import (
	"strconv"

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

// DefaultClassMapper is the default error classification mapper.
var DefaultClassMapper = &ClassMapper{
	Majors: map[errors.Major]ErrFunc{
		errors.MjrInternal: ErrInternalError,
		codec.MjrCodec:     ErrInvalidInput,
		filter.MjrFilter:   ErrInvalidQueryParameter,
	},
	Minors: map[errors.Minor]ErrFunc{
		codec.MnrUnmarshal:    ErrInvalidInput,
		query.MnrViolation:    ErrInvalidJSONFieldValue,
		mapping.MnrModel:      ErrInternalError,
		mapping.MnrRepository: ErrInternalError,
	},
	Class: map[errors.Class]ErrFunc{
		// Query:
		query.ClassFieldValue:         ErrInvalidJSONFieldValue,
		query.ClassInvalidParameter:   ErrInvalidQueryParameter,
		filter.ClassFilterCollection:  ErrInvalidResourceName,
		query.ClassInvalidSort:        ErrInvalidQueryParameter,
		query.ClassInvalidFieldSet:    ErrInvalidQueryParameter,
		query.ClassNoFieldsInFieldSet: ErrInvalidQueryParameter,
		query.ClassViolationUnique:    ErrResourceAlreadyExists,
		query.ClassViolationCheck:     ErrInvalidJSONFieldValue,
		query.ClassViolationNotNull:   ErrInvalidInput,
		codec.ClassMarshal:            ErrInternalError,
		// Codec:
		codec.ClassUnmarshalDocument: ErrInvalidJSONDocument,
		// Mapping:
		mapping.ClassFieldValue:           ErrInvalidJSONFieldValue,
		mapping.ClassFieldNotParser:       ErrInvalidQueryParameter,
		mapping.ClassFieldNotNullable:     ErrForbiddenOperation,
		mapping.ClassModelNotFound:        ErrResourceNotFound,
		mapping.ClassInvalidRelationField: ErrInvalidQueryParameter,
		// Server:
		server.ClassNotAcceptable:         ErrNotAcceptable,
		server.ClassUnsupportedHeader:     ErrUnsupportedHeader,
		server.ClassMissingRequiredHeader: ErrMissingRequiredHeader,
		server.ClassHeaderValue:           ErrInvalidHeaderValue,
		// Auth:
		auth.ClassAccountNotFound:     ErrInvalidAuthenticationInfo,
		auth.ClassInvalidSecret:       ErrInvalidAuthenticationInfo,
		auth.ClassForbidden:           ErrForbiddenAuthorize,
		auth.ClassInvalidRole:         ErrForbiddenAuthorize,
		auth.ClassAuthorizationHeader: ErrInvalidAuthorizationHeader,
	},
}

// MapError maps the 'err' input error into slice of 'Error'.
// The function uses DefaultClassMapper for error mapping.
// The logic is the same as for DefaultClassMapper.Errors method.
func MapError(err error) []*codec.Error {
	return DefaultClassMapper.errors(err)
}

// ClassMapper is the neuron errors classification mapper.
// It creates the 'Error' from the provided error.
type ClassMapper struct {
	Majors map[errors.Major]ErrFunc
	Minors map[errors.Minor]ErrFunc
	Class  map[errors.Class]ErrFunc
}

// Errors gets the slice of 'Error' from the provided 'err' error.
// The mapping is based on the 'most specific classification first' method.
// If the error is 'errors.ClassError' the function gets it's class.
// The function checks classification occurrence based on the priority:
//	- Class
//	- Minor
//	- Major
// If no mapping is provided for given classification - an internal error is returned.
func (c *ClassMapper) Errors(err error) []*codec.Error {
	return c.errors(err)
}

func (c *ClassMapper) errors(err error) []*codec.Error {
	switch et := err.(type) {
	case *codec.Error:
		return []*codec.Error{et}
	case codec.MultiError:
		return et
	case errors.ClassError:
		return []*codec.Error{c.mapSingleError(et)}
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

func (c *ClassMapper) mapSingleError(e errors.ClassError) *codec.Error {
	// check if the class is stored in the mapper
	creator, ok := c.Class[e.Class()]
	if !ok {
		// otherwise check it's minor
		creator, ok = c.Minors[e.Class().Minor()]
		if !ok {
			// at last check it's major
			creator, ok = c.Majors[e.Class().Major()]
			if !ok {
				log.Errorf("Unmapped error provided: %v, with Class: %v", e, e.Class())
				return ErrInternalError()
			}
		}
	}

	err := creator()
	err.Code = strconv.FormatInt(int64(e.Class()), 16)
	detailed, ok := e.(*errors.DetailedError)
	if ok {
		err.Detail = detailed.Details
		err.ID = detailed.ID.String()
	}
	return err
}
