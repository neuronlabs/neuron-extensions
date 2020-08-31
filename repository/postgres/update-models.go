package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
)

// UpdateModels implements repository.Repository interface. Updates input models in the query.
func (p *Postgres) UpdateModels(ctx context.Context, s *query.Scope) (affected int64, err error) {
	switch len(s.FieldSets) {
	case 0:
		return 0, errors.Wrap(query.ErrInvalidFieldSet, "no fields to update")
	case 1:
		fieldSet := s.FieldSets[0]
		switch len(s.Models) {
		case 0:
			return 0, errors.Wrap(query.ErrNoModels, "no models to update")
		case 1:
			model := s.Models[0]
			return p.updatedModelWithFieldset(ctx, s, fieldSet, model)
		}
		b := &pgx.Batch{}
		if err := p.updateBatchModelsWithFieldSet(s, b, fieldSet, s.Models...); err != nil {
			return 0, err
		}

		results := p.connection(s).SendBatch(ctx, b)
		defer results.Close()
		for i := 0; i < b.Len(); i++ {
			tag, err := results.Exec()
			if err != nil {
				return affected, errors.Wrap(p.neuronError(err), err.Error())
			}
			affected += tag.RowsAffected()
		}
		return affected, nil
	default:
		return p.updateModelsWithBulkFieldSet(ctx, s)
	}
}

func (p *Postgres) updateModelsWithBulkFieldSet(ctx context.Context, s *query.Scope) (affected int64, err error) {
	var models []mapping.Model
	b := &pgx.Batch{}
	// For each unique fieldset create a query that would be executed for each matched model.
	// This would result in a query for each model.
	bulk := &mapping.BulkFieldSet{}
	for i, fieldSet := range s.FieldSets {
		bulk.Add(fieldSet, i)
	}

	for _, fieldSet := range bulk.FieldSets {
		indices := bulk.GetIndicesByFieldset(fieldSet)
		for _, index := range indices {
			models = append(models, s.Models[index])
		}
		if err = p.updateBatchModelsWithFieldSet(s, b, fieldSet, models...); err != nil {
			if !errors.Is(err, query.ErrNoFieldsInFieldSet) {
				return affected, err
			}
		}
		internal.ResetIncrementor(s)
	}

	results := p.connection(s).SendBatch(ctx, b)
	defer results.Close()
	for i := 0; i < b.Len(); i++ {
		tag, err := results.Exec()
		if err != nil {
			return affected, errors.Wrap(p.neuronError(err), err.Error())
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
		return 0, errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.ModelStruct)
	}

	var (
		modelValues []interface{}
	)
	primaryValue := model.GetPrimaryKeyValue()
	for _, field := range fieldSet {
		if field.DatabaseNotNull() && field.Kind() == mapping.KindForeignKey {
			isZero, err := fielder.IsFieldZero(field)
			if err != nil {
				return 0, err
			}
			if isZero {
				modelValues = append(modelValues, nil)
				continue
			}
		}
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
		return affected, errors.WrapDetf(p.neuronError(err), "update failed: %v", err)
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
			return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.ModelStruct)
		}
		var (
			modelValues []interface{}
		)
		primaryValue := model.GetPrimaryKeyValue()
		for _, field := range fieldSet {
			if field.DatabaseNotNull() && field.Kind() == mapping.KindForeignKey {
				isZero, err := fielder.IsFieldZero(field)
				if err != nil {
					return err
				}
				if isZero {
					modelValues = append(modelValues, nil)
					continue
				}
			}
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
