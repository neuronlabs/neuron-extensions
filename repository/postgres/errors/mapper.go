package errors

import (
	"github.com/jackc/pgconn"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/repository"
)

var pqMapping = map[string]errors.Class{
	// Class 02 - No data
	"02":    query.ClassNoResult,
	"P0002": query.ClassNoResult,

	// Class 08 - Connection Exception
	"08": repository.ClassConnection,

	"0B000": query.ClassTxState,

	// Class 21 - Cardinality Violation
	"21": query.ClassInternal,

	// Class 22 Data Exception
	"22": query.ClassFieldValue,

	// Class 23 Integrity Violation errors
	"23":    query.ClassViolationIntegrityConstraint,
	"23000": query.ClassViolationIntegrityConstraint,
	"23001": query.ClassViolationRestrict,
	"23502": query.ClassViolationNotNull,
	"23503": query.ClassViolationForeignKey,
	"23505": query.ClassViolationUnique,
	"23514": query.ClassViolationCheck,

	// Class 25 Invalid Transaction State
	"25": query.ClassTxState,

	// Class 28 Invalid Authorization Specification
	"28000": repository.ClassAuthorization,
	"28P01": repository.ClassAuthorization,

	// Class 2D Invalid Transaction Termination
	"2D000": query.ClassTxState,

	"3D": query.ClassInternal,

	// Class 3F Invalid Schema Name
	"3F":    query.ClassInternal,
	"3F000": query.ClassInternal,

	// Class 40 - Transaction Rollback
	"40": query.ClassTxState,

	// Class 42 - Invalid Syntax
	"42":    query.ClassInternal,
	"42939": repository.ClassReservedName,
	"42804": query.ClassViolationDataType,
	"42703": query.ClassInternal,
	"42883": query.ClassInternal,
	"42P01": query.ClassInternal,
	"42701": query.ClassInternal,
	"42P06": query.ClassInternal,
	"42P07": query.ClassInternal,
	"42501": repository.ClassAuthorization,

	// Class 53 - Insufficient Resources
	"53": repository.ClassService,

	// Class 54 - Program Limit Exceeded
	"54": repository.ClassService,

	// Class 58 - System Errors
	"58": repository.ClassService,

	// Class XX - Internal Error
	"XX": repository.ClassService,
}

// Get gets the mapped postgres pq.Error to the neuron error class.
func Get(err error) (errors.Class, bool) {
	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return ClassInternal, false
	}

	cl, ok := pqMapping[pgErr.Code]
	if ok {
		return cl, ok
	}
	if len(pgErr.Code) >= 2 {
		cl, ok = pqMapping[pgErr.Code[0:2]]
		return cl, ok
	}
	return ClassInternal, ok
}
