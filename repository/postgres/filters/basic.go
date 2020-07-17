package filters

import (
	"strings"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/internal"
	"github.com/neuronlabs/neuron-plugins/repository/postgres/migrate"
)

// BasicSQLizer gets the SQLQueries from the provided filter.
func BasicSQLizer(s *query.Scope, quotedWriter internal.QuotedWordWriteFunc, field *mapping.StructField, fv *query.OperatorValues) (SQLQueries, error) {
	queries := SQLQueries{}

	columnName, err := migrate.FieldColumnName(field)
	if err != nil {
		return nil, err
	}

	op, err := getSQLOperator(fv.Operator)
	if err != nil {
		return nil, err
	}
	b := &strings.Builder{}
	for _, v := range fv.Values {
		quotedWriter(b, columnName)
		b.WriteString(" ")
		b.WriteString(op)
		b.WriteString(" ")
		b.WriteString(internal.StringIncrementor(s))
		queries = append(queries, &SQLQuery{Query: b.String(), Values: []interface{}{v}})
		b.Reset()
	}
	return queries, nil
}

// NullSQLizer is the SQLizer function that returns NULL'ed queries.
func NullSQLizer(s *query.Scope, quotedWriter internal.QuotedWordWriteFunc, field *mapping.StructField, fv *query.OperatorValues) (SQLQueries, error) {
	columnName, err := migrate.FieldColumnName(field)
	if err != nil {
		return nil, err
	}

	op, err := getSQLOperator(fv.Operator)
	if err != nil {
		return nil, err
	}

	b := &strings.Builder{}
	quotedWriter(b, columnName)
	b.WriteString(" ")
	b.WriteString(op)

	queries := SQLQueries{&SQLQuery{Query: b.String()}}

	return queries, nil
}

// InSQLizer creates the SQLQueries for the 'IN' and 'NOT IN' filter Operators.
func InSQLizer(s *query.Scope, quotedWriter internal.QuotedWordWriteFunc, field *mapping.StructField, fv *query.OperatorValues) (SQLQueries, error) {
	if fv.Values == nil || len(fv.Values) == 0 {
		return SQLQueries{}, nil
	}
	columnName, err := migrate.FieldColumnName(field)
	if err != nil {
		return nil, err
	}

	op, err := getSQLOperator(fv.Operator)
	if err != nil {
		return nil, err
	}

	b := &strings.Builder{}

	quotedWriter(b, columnName)
	b.WriteString(" ")
	b.WriteString(op)
	b.WriteString(" (")

	for i := range fv.Values {
		b.WriteString(internal.StringIncrementor(s))
		if i != len(fv.Values)-1 {
			b.WriteRune(',')
		}
	}
	b.WriteRune(')')

	queries := SQLQueries{&SQLQuery{Query: b.String(), Values: fv.Values}}

	return queries, nil
}

// StringOperatorsSQLizer creates the SQLQueries for the provided filter values.
func StringOperatorsSQLizer(s *query.Scope, quotedWriter internal.QuotedWordWriteFunc, field *mapping.StructField, fv *query.OperatorValues) (SQLQueries, error) {
	columnName, err := migrate.FieldColumnName(field)
	if err != nil {
		return nil, err
	}

	op, err := getSQLOperator(fv.Operator)
	if err != nil {
		return nil, err
	}

	queries := SQLQueries{}

	b := &strings.Builder{}
	for _, v := range fv.Values {
		strValue, ok := v.(string)
		if !ok {
			return nil, errors.NewDetf(query.ClassFilterValues, "operator: '%s' requires string filter values", fv.Operator.Name)
		}

		switch fv.Operator {
		case query.OpStartsWith:
			strValue += "%"
		case query.OpEndsWith:
			strValue = "%" + strValue
		case query.OpContains:
			strValue = "%" + strValue + "%"
		}

		quotedWriter(b, columnName)
		b.WriteRune(' ')
		b.WriteString(op)
		b.WriteRune(' ')
		b.WriteString(internal.StringIncrementor(s))

		queries = append(queries, &SQLQuery{Query: b.String(), Values: []interface{}{strValue}})
		b.Reset()
	}

	return queries, nil
}
