package errors

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// MjrPostgres is the major error classification for the postgres repository.
	MjrPostgres errors.Major

	// ClassUnmappedError is the error classification for unmapped errors.
	ClassUnmappedError errors.Class

	// ClassInternal is the internal error in the postgres repository package.
	ClassInternal errors.Class
)

func init() {
	MjrPostgres = errors.MustNewMajor()
	ClassUnmappedError = errors.MustNewMajorClass(MjrPostgres)
	ClassInternal = errors.MustNewMajorClass(errors.MjrInternal)
}
