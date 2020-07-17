package postgres

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/filters"
	"github.com/neuronlabs/neuron-plugins/repository/postgres/migrate"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/internal"
)

var _ query.Updater = &Postgres{}

// Update patches all the values that matches scope's filters, sorts and pagination
// Implements query.Updater interface
func (p *Postgres) Update(ctx context.Context, s *query.Scope) (int64, error) {
	// There are two possibilities - update with filters or update models.
	// The first one must contain a single model and the filters.
	// Whereas the second one must contain models with non zero primary field value.
	if len(s.Filters) != 0 {
		return p.updateWithFilters(ctx, s)
	}
	return p.updateModels(ctx, s)
}

func (p *Postgres) updateModels(ctx context.Context, s *query.Scope) (affected int64, err error) {
	if s.BulkFieldSets != nil {
		return p.updateModelsWithBulkFieldSet(ctx, s)
	}
	switch len(s.Models) {
	case 0:
		return 0, errors.New(query.ClassNoModels, "no models to update")
	case 1:
		model := s.Models[0]
		// Check if this is about to update all models.
		if model.IsPrimaryKeyZero() {
			return p.updateWithFilters(ctx, s)
		}
		return p.updatedModelWithFieldset(ctx, s, s.FieldSet, s.Models[0])
	}

	b := &pgx.Batch{}
	if err := p.updateBatchModelsWithFieldSet(s, b, s.FieldSet, s.Models...); err != nil {
		return 0, err
	}

	results := p.connection(s).SendBatch(ctx, b)
	for i := 0; i < b.Len(); i++ {
		tag, err := results.Exec()
		if err != nil {
			return affected, err
		}
		affected += tag.RowsAffected()
	}
	return affected, nil

}

func (p *Postgres) updateModelsWithBulkFieldSet(ctx context.Context, s *query.Scope) (affected int64, err error) {
	var models []mapping.Model
	b := &pgx.Batch{}
	// For each unique fieldset create a query that would be executed for each matched model.
	// This would result in a query for each model.
	for _, fieldSet := range s.BulkFieldSets.FieldSets {
		indices := s.BulkFieldSets.GetIndicesByFieldset(fieldSet)
		for _, index := range indices {
			models = append(models, s.Models[index])
		}
		if err = p.updateBatchModelsWithFieldSet(s, b, fieldSet, models...); err != nil {
			if !errors.IsClass(err, query.ClassNoFieldsInFieldSet) {
				return affected, err
			}
		}
		internal.ResetIncrementor(s)
	}

	results := p.connection(s).SendBatch(ctx, b)
	for i := 0; i < b.Len(); i++ {
		tag, err := results.Exec()
		if err != nil {
			return affected, err
		}
		affected += tag.RowsAffected()
	}
	return affected, nil
}

func (p *Postgres) updatedModelWithFieldset(ctx context.Context, s *query.Scope, fieldSet mapping.FieldSet, model mapping.Model) (affected int64, err error) {
	fieldSet, err = p.prepareUpdateModelFieldSet(fieldSet)
	if err != nil {
		return 0, err
	}

	q, err := p.buildUpdateModelQuery(s, fieldSet)
	if err != nil {
		return 0, err
	}
	fielder, ok := model.(mapping.Fielder)
	if !ok {
		return 0, errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.ModelStruct)
	}

	var (
		modelValues []interface{}
	)
	primaryValue := model.GetPrimaryKeyValue()
	for _, field := range fieldSet {
		fieldValue, err := fielder.GetFieldValue(field)
		if err != nil {
			return affected, err
		}
		modelValues = append(modelValues, fieldValue)
	}

	// Primary key value must be the last one - it would be set as the filter value.
	modelValues = append(modelValues, primaryValue)

	tag, err := p.connection(s).Exec(ctx, q, modelValues...)
	if err != nil {
		return affected, err
	}

	return tag.RowsAffected(), nil
}

func (p *Postgres) updateBatchModelsWithFieldSet(s *query.Scope, b internal.Batch, fieldSet mapping.FieldSet, models ...mapping.Model) (err error) {
	fieldSet, err = p.prepareUpdateModelFieldSet(fieldSet)
	if err != nil {
		return err
	}

	q, err := p.buildUpdateModelQuery(s, fieldSet)
	if err != nil {
		return err
	}

	for _, model := range models {
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.ModelStruct)
		}
		var (
			modelValues []interface{}
		)
		primaryValue := model.GetPrimaryKeyValue()
		for _, field := range fieldSet {
			fieldValue, err := fielder.GetFieldValue(field)
			if err != nil {
				return err
			}
			modelValues = append(modelValues, fieldValue)
		}
		// Primary key value must be the last one - it would be set as the filter value.
		modelValues = append(modelValues, primaryValue)

		b.Queue(q, modelValues...)
	}
	return nil
}

func (p *Postgres) buildUpdateModelQuery(s *query.Scope, fieldSet mapping.FieldSet) (string, error) {
	sb := &strings.Builder{}
	if err := p.buildUpdateQuery(s, fieldSet, sb); err != nil {
		return "", err
	}
	sb.WriteString(" WHERE ")
	primary, err := migrate.FieldColumnName(s.ModelStruct.Primary())
	if err != nil {
		return "", err
	}
	sb.WriteString(primary)
	sb.WriteString(" = $")
	sb.WriteString(strconv.Itoa(internal.Incrementor(s)))
	q := sb.String()
	return q, nil
}

func (p *Postgres) buildUpdateQuery(s *query.Scope, fieldSet mapping.FieldSet, sb *strings.Builder) error {
	table, err := migrate.ModelsTable(s.ModelStruct)
	if err != nil {
		return err
	}
	sb.WriteString("UPDATE ")
	p.writeQuotedWord(sb, table.Schema)
	sb.WriteRune('.')
	p.writeQuotedWord(sb, table.Name)
	sb.WriteString(" SET ")

	for i, field := range fieldSet {
		column, err := migrate.FieldsColumn(field)
		if err != nil {
			return err
		}
		sb.WriteString(column.Name)
		sb.WriteString(" = ")
		sb.WriteRune('$')
		sb.WriteString(strconv.Itoa(internal.Incrementor(s)))
		if i != len(fieldSet)-1 {
			sb.WriteString(", ")
		}
	}
	return nil
}

func (p *Postgres) updateWithFilters(ctx context.Context, s *query.Scope) (int64, error) {
	// Check if there is anything to update.
	if len(s.FieldSet) == 0 {
		return 0, errors.New(query.ClassInvalidFieldSet, "provided empty fieldset - update with filters")
	}
	// Check if there is exactly one model.
	if len(s.Models) != 1 {
		return 0, errors.New(query.ClassInvalidModels, "update with filters require exactly one model")
	}

	sb := &strings.Builder{}
	// Build update query.
	if err := p.buildUpdateQuery(s, s.FieldSet, sb); err != nil {
		return 0, err
	}

	// Get model fielder and get it's fields values.
	var values []interface{}
	fielder, ok := s.Models[0].(mapping.Fielder)
	if !ok {
		return 0, errors.New(mapping.ClassModelNotImplements, "model doesn't implement Fielder interface")
	}

	for _, field := range s.FieldSet {
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
		return 0, err
	}
	return tag.RowsAffected(), nil
}
