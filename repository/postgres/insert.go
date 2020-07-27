package postgres

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/migrate"
)

// Insert depending on the query efficiently inserts models with related fieldSets.
// Implements repository.Repository interface.
func (p *Postgres) Insert(ctx context.Context, s *query.Scope) error {
	if len(s.FieldSets) == 1 {
		return p.insertWithCommonFieldSet(ctx, s)
	}
	return p.insertWithBulkFieldSet(ctx, s)
}

//
// PRIVATE
//

func (p *Postgres) insertWithCommonFieldSet(ctx context.Context, s *query.Scope) error {
	q, err := p.parseInsertWithCommonFieldSet(s)
	if err != nil {
		return err
	}
	if q.primarySelected {
		_, err := p.connection(s).Exec(ctx, q.query, q.values...)
		if err != nil {
			return errors.NewDetf(p.errorClass(err), "inserting failed: %v", err)
		}
		return nil
	}

	rows, err := p.connection(s).Query(ctx, q.query, q.values...)
	if err != nil {
		return errors.NewDetf(p.errorClass(err), "creating query failed")
	}
	var i int
	for rows.Next() {
		if err = rows.Scan(s.Models[i].GetPrimaryKeyAddress()); err != nil {
			return errors.NewDetf(p.errorClass(err), "inserting failed: %v", err)
		}
		i++
	}
	rows.Close()
	return nil
}

func (p *Postgres) insertWithBulkFieldSet(ctx context.Context, s *query.Scope) error {
	b := &pgx.Batch{}
	q, err := p.parseInsertBulkFieldsetQuery(s, b)
	if err != nil {
		return err
	}

	br := p.connection(s).SendBatch(ctx, b)
	defer br.Close()

	for _, indices := range q {
		switch len(indices) {
		case 0:
			if _, err = br.Exec(); err != nil {
				return err
			}
		default:
			rows, err := br.Query()
			if err != nil {
				return err
			}

			var i int
			for rows.Next() {
				if err = rows.Scan(s.Models[indices[i]].GetPrimaryKeyAddress()); err != nil {
					rows.Close()
					return err
				}
				i++
			}
			rows.Close()
		}
	}
	return nil
}

type insertQuery struct {
	query           string
	values          []interface{}
	primarySelected bool
}

func (p *Postgres) parseInsertWithCommonFieldSet(s *query.Scope) (*insertQuery, error) {
	// Models length is already checked - must be one.
	t, err := migrate.ModelsTable(s.ModelStruct)
	if err != nil {
		return nil, err
	}

	commonFieldSet, hasCommonFieldSet := s.CommonFieldSet()
	if !hasCommonFieldSet {
		return nil, errors.NewDetf(query.ClassInvalidFieldSet, "no insert fieldset provided")
	}
	fieldSet, autoSelected := p.prepareInsertFieldset(s.ModelStruct, commonFieldSet)

	iq := &insertQuery{}
	sb := &strings.Builder{}
	// Build the query of form "INSERT INTO schemaName.tableName (fields) VALUES (fieldValues)"
	sb.WriteString("INSERT INTO ")
	p.writeQuotedWord(sb, t.Schema)
	sb.WriteRune('.')
	p.writeQuotedWord(sb, t.Name)

	if len(fieldSet) > 0 {
		sb.WriteString(" (")
		for i, field := range fieldSet {
			if field.Kind() == mapping.KindPrimary {
				iq.primarySelected = true
			}
			dbName, err := migrate.FieldColumnName(field)
			if err != nil {
				return nil, err
			}
			p.writeQuotedWord(sb, dbName)
			if i != len(fieldSet)-1 {
				sb.WriteRune(',')
			}
		}
		sb.WriteString(") VALUES ")
		for j, model := range s.Models {
			sb.WriteRune('(')
			// Get the model and get selected field values.
			fielder, isFielder := model.(mapping.Fielder)
			if !isFielder && (len(fieldSet) > 1 || ((len(fieldSet) == 1) && fieldSet[0].Kind() != mapping.KindPrimary)) {
				return nil, errors.Newf(mapping.ClassModelNotImplements, "Model: '%s' doesn't implement Fielder interface", s.ModelStruct)
			}

			// Add the selected fields to the query.
			var fieldValue interface{}
			for i, field := range fieldSet {
				switch field.Kind() {
				case mapping.KindPrimary:
					iq.values = append(iq.values, model.GetPrimaryKeyValue())
				default:
					if autoSelected != nil && autoSelected.Contains(field) {
						fieldValue, err = fielder.GetFieldZeroValue(field)
					} else {
						fieldValue, err = fielder.GetFieldValue(field)
					}
					if err != nil {
						return nil, err
					}
					iq.values = append(iq.values, fieldValue)
				}

				// Write value string incrementor.
				sb.WriteRune('$')
				sb.WriteString(strconv.Itoa(internal.Incrementor(s)))

				// Add comma if there is more fields to write
				if i != len(fieldSet)-1 {
					sb.WriteRune(',')
				}
			}
			sb.WriteRune(')')
			if j != len(s.Models)-1 {
				sb.WriteRune(',')
			}
		}
	} else {
		sb.WriteString(" VALUES ")
		for i := range s.Models {
			sb.WriteString("(DEFAULT)")
			if i != len(s.Models)-1 {
				sb.WriteRune(',')
			}
		}
	}
	if !iq.primarySelected {
		primaryDBName, err := migrate.FieldColumnName(s.ModelStruct.Primary())
		if err != nil {
			return nil, err
		}
		sb.WriteString(" RETURNING ")
		p.writeQuotedWord(sb, primaryDBName)
	}
	iq.query = sb.String()
	return iq, nil
}

// parseInsertBulkFieldSetQuery prepares the string query with the bulk fieldset for provided models.
func (p *Postgres) parseInsertBulkFieldsetQuery(s *query.Scope, batch internal.Batch) (queryIndices [][]int, err error) {
	t, err := migrate.ModelsTable(s.ModelStruct)
	if err != nil {
		return nil, err
	}

	primaryKeyName, err := migrate.FieldColumnName(s.ModelStruct.Primary())
	if err != nil {
		return nil, err
	}
	primaryKeyName = migrate.GetQuotedWord(primaryKeyName, p.PostgresVersion)

	var (
		sb           strings.Builder
		autoSelected mapping.FieldSet
	)

	bulk := &mapping.BulkFieldSet{}
	for i, fieldSet := range s.FieldSets {
		bulk.Add(fieldSet, i)
	}

	queryIndices = make([][]int, len(bulk.FieldSets))
	for i := range bulk.FieldSets {
		var values []interface{}
		sb.WriteString("INSERT INTO ")
		p.writeQuotedWord(&sb, t.Schema)
		sb.WriteRune('.')
		p.writeQuotedWord(&sb, t.Name)

		// Get the fieldset and related model indices, add to the query indices and trim the fieldset.
		fieldSet := bulk.FieldSets[i]
		indices := bulk.GetIndicesByFieldset(fieldSet)
		fieldSet, autoSelected = p.prepareInsertFieldset(s.ModelStruct, fieldSet)

		var primarySelected bool
		// Write fieldset column names (id, name, surname).
		if len(fieldSet) > 0 {
			sb.WriteString(" (")
			for j, field := range fieldSet {
				if field.Kind() == mapping.KindPrimary {
					primarySelected = true
				}
				dbName, err := migrate.FieldColumnName(field)
				if err != nil {
					return nil, err
				}
				p.writeQuotedWord(&sb, dbName)
				if j != len(fieldSet)-1 {
					sb.WriteRune(',')
				}
			}
			sb.WriteRune(')')
		}
		sb.WriteString(" VALUES ")

		// Write comma separated, wrapped in brackets field value for given models i.e. ($1,$2,$3),($4,$5,$6).
		for j, index := range indices {
			sb.WriteRune('(')
			if len(fieldSet) != 0 {
				model := s.Models[index]
				fielder, isFielder := model.(mapping.Fielder)
				if !isFielder && (len(fieldSet) > 1 || ((len(fieldSet) == 1) && fieldSet[0].Kind() != mapping.KindPrimary)) {
					return nil, errors.Newf(mapping.ClassModelNotImplements, "Model: '%s' doesn't implement Fielder interface", s.ModelStruct)
				}

				var fieldValue interface{}
				for k, field := range fieldSet {
					switch field.Kind() {
					case mapping.KindPrimary:
						values = append(values, model.GetPrimaryKeyValue())
					default:
						if autoSelected != nil && autoSelected.Contains(field) {
							fieldValue, err = fielder.GetFieldZeroValue(field)
						} else {
							fieldValue, err = fielder.GetFieldValue(field)
						}
						if err != nil {
							return nil, err
						}
						values = append(values, fieldValue)
					}

					sb.WriteRune('$')
					sb.WriteString(strconv.Itoa(internal.Incrementor(s)))
					if k != len(fieldSet)-1 {
						sb.WriteRune(',')
					}
				}
			} else {
				sb.WriteString("DEFAULT")
			}
			sb.WriteRune(')')
			if j != len(indices)-1 {
				sb.WriteRune(',')
			}
		}

		if !primarySelected {
			sb.WriteString(" RETURNING ")
			sb.WriteString(primaryKeyName)
			queryIndices[i] = indices
		}
		batch.Queue(sb.String(), values...)
		sb.Reset()
		internal.ResetIncrementor(s)
	}
	return queryIndices, nil
}
