package postgres

import (
	"context"
	"strconv"
	"strings"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/filters"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
)

// Update patches all the values that matches scope's filters, sorts and pagination
// Implements repository.Repository interface
func (p *Postgres) Update(ctx context.Context, s *query.Scope) (int64, error) {
	// Check if there is anything to update.
	if len(s.FieldSets) != 1 {
		return 0, errors.Wrap(query.ErrInvalidFieldSet, "provided empty fieldset length - update with filters")
	}

	fieldSet := s.FieldSets[0]
	if len(fieldSet) == 0 {
		return 0, errors.Wrap(query.ErrInvalidFieldSet, "provided empty fieldset - update with filters")
	}

	// Check if there is exactly one model.
	if len(s.Models) != 1 {
		return 0, errors.Wrap(query.ErrInvalidModels, "update with filters require exactly one model")
	}

	// Build update query.
	sb := &strings.Builder{}
	if err := p.buildUpdateQuery(s, fieldSet, sb); err != nil {
		return 0, err
	}

	// Get model fielder and get it's fields values.
	var values []interface{}
	fielder, ok := s.Models[0].(mapping.Fielder)
	if !ok {
		return 0, errors.Wrap(mapping.ErrModelNotImplements, "model doesn't implement Fielder interface")
	}

	for _, field := range fieldSet {
		fieldValue, err := fielder.GetFieldValue(field)
		if err != nil {
			return 0, err
		}
		values = append(values, fieldValue)
	}

	// Parse filters and store in the string builder.
	parsedFilters, err := filters.ParseFilters(s, p.writeQuotedWord)
	if err != nil {
		return 0, err
	}

	if len(parsedFilters) > 0 {
		sb.WriteString(" WHERE ")
		for i, f := range parsedFilters {
			sb.WriteString(f.Query)
			if i < len(parsedFilters)-1 {
				sb.WriteString(" AND ")
			}
			values = append(values, f.Values...)
		}
	}

	tag, err := p.connection(s).Exec(ctx, sb.String(), values...)
	if err != nil {
		return 0, errors.WrapDetf(p.neuronError(err), "update failed: %v", err)
	}
	return tag.RowsAffected(), nil
}

func (p *Postgres) buildUpdateModelQuery(s *query.Scope, fieldSet mapping.FieldSet) (string, error) {
	sb := &strings.Builder{}
	if err := p.buildUpdateQuery(s, fieldSet, sb); err != nil {
		return "", err
	}
	sb.WriteString(" WHERE ")
	sb.WriteString(s.ModelStruct.Primary().DatabaseName)
	sb.WriteString(" = $")
	sb.WriteString(strconv.Itoa(internal.Incrementor(s)))
	q := sb.String()
	return q, nil
}

func (p *Postgres) buildUpdateQuery(s *query.Scope, fieldSet mapping.FieldSet, sb *strings.Builder) error {
	sb.WriteString("UPDATE ")
	p.writeQuotedWord(sb, s.ModelStruct.DatabaseSchemaName)
	sb.WriteRune('.')
	p.writeQuotedWord(sb, s.ModelStruct.DatabaseName)
	sb.WriteString(" SET ")

	for i, field := range fieldSet {
		sb.WriteString(field.DatabaseName)
		sb.WriteString(" = ")
		sb.WriteRune('$')
		sb.WriteString(strconv.Itoa(internal.Incrementor(s)))
		if i != len(fieldSet)-1 {
			sb.WriteString(", ")
		}
	}
	return nil
}
