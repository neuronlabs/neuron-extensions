package filters

import (
	"github.com/neuronlabs/neuron-plugins/repository/postgres/internal"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/log"
)

func init() {
	ClassInternal = errors.MustNewMajorClass(controller.MjrInternal)
}

var (
	operatorSQL      []string
	operatorSQLizers []SQLizer
	// ClassInternal defines internal filters error.
	ClassInternal errors.Class
)

func init() {
	operatorSQL = make([]string, 16)
	operatorSQLizers = make([]SQLizer, 16)
	registerOperator(query.OpEqual, BasicSQLizer, "=")
	registerOperator(query.OpIn, InSQLizer, "IN")
	registerOperator(query.OpNotIn, InSQLizer, "NOT IN")
	registerOperator(query.OpNotEqual, BasicSQLizer, "<>")
	registerOperator(query.OpGreaterEqual, BasicSQLizer, ">=")
	registerOperator(query.OpGreaterThan, BasicSQLizer, ">")
	registerOperator(query.OpLessEqual, BasicSQLizer, "<=")
	registerOperator(query.OpLessThan, BasicSQLizer, "<")
	registerOperator(query.OpContains, StringOperatorsSQLizer, "LIKE")
	registerOperator(query.OpStartsWith, StringOperatorsSQLizer, "LIKE")
	registerOperator(query.OpEndsWith, StringOperatorsSQLizer, "LIKE")
	registerOperator(query.OpIsNull, NullSQLizer, "IS NULL")
	registerOperator(query.OpNotNull, NullSQLizer, "IS NOT NULL")
}

// SQLQuery defines the SQL query Models pair
type SQLQuery struct {
	Query  string
	Values []interface{}
}

// SQLQueries is the wrapper arount the SQL queries value.
type SQLQueries []*SQLQuery

// SQLizer is the function that sqlizes provided OperatorValuePair.
type SQLizer func(*query.Scope, internal.QuotedWordWriteFunc, *mapping.StructField, *query.OperatorValues) (SQLQueries, error)

// SQLOperator gets the operator sql name.
func SQLOperator(o *query.Operator) (string, error) {
	return getSQLOperator(o)
}

// RegisterSQLizer registers new SQLizer function for the provided operator. Optionally can set the raw SQL value.
func RegisterSQLizer(o *query.Operator, sqlizer SQLizer, raw ...string) {
	registerOperator(o, sqlizer, raw...)
}

/** PRIVATE */

func getSQLOperator(o *query.Operator) (string, error) {
	if o == nil {
		log.Errorf("Provided nil filter operator.")
		return "", errors.NewDet(query.ClassFilterFormat, "provided nil operator")
	}
	if int(o.ID) > len(operatorSQL)-1 {
		log.Errorf("Cannot get filter operator: '%s' SQL ", o.Name)
		return "", errors.NewDet(query.ClassFilterFormat, "unsupported filter operator")
	}

	sql := operatorSQL[o.ID]
	if sql == "" {
		log.Errorf("Operator: '%s' has SQL value ", o.Name)
		return "", errors.NewDetf(query.ClassFilterFormat, "filter operator: '%v' hase no SQL value", o.Name)
	}

	return sql, nil
}

func getOperatorSQLizer(o *query.Operator) (SQLizer, error) {
	if int(o.ID) > len(operatorSQLizers)-1 {
		return nil, errors.NewDet(query.ClassFilterFormat, "unsupported filter operator")
	}

	return operatorSQLizers[o.ID], nil
}

func registerOperator(o *query.Operator, sqlizer SQLizer, raw ...string) {
	registerOperatorSQLizer(o, sqlizer)
	if len(raw) > 0 {
		registerOperatorRawSQL(o, raw[0])
	}
}

func registerOperatorSQLizer(o *query.Operator, sqlizer SQLizer) {
	minSize := len(operatorSQLizers) - 1

	for int(o.ID) > minSize {
		if minSize == 0 {
			minSize = 1
		}
		minSize *= 2
	}

	if minSize != len(operatorSQLizers)-1 {
		temp := make([]SQLizer, minSize)
		copy(temp, operatorSQLizers)
		operatorSQLizers = temp
	}

	operatorSQLizers[o.ID] = sqlizer
}

func registerOperatorRawSQL(o *query.Operator, raw string) {
	minSize := len(operatorSQL) - 1

	for int(o.ID) > minSize {
		if minSize == 0 {
			minSize = 1
		}
		minSize *= 2
	}

	if minSize != len(operatorSQL)-1 {
		temp := make([]string, minSize)
		copy(temp, operatorSQL)
		operatorSQL = temp
	}

	operatorSQL[o.ID] = raw
}
