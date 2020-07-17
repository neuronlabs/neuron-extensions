package migrate

import (
	"context"
	"fmt"
	"strings"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/internal"
	"github.com/neuronlabs/neuron-plugins/repository/postgres/log"
)

// Table is the model for the SQL table definition.
type Table struct {
	Schema       string
	QuotedSchema string
	Name         string
	QuotedName   string
	Columns      []*Column
	Indexes      []*Index

	// private
	model *mapping.ModelStruct
}

// Model is the *mapping.ModelStruct for which the Table was defined.
func (t *Table) Model() *mapping.ModelStruct {
	return t.model
}

// FindIndex finds the index by it's name.
func (t *Table) FindIndex(name string) *Index {
	for _, i := range t.Indexes {
		if i.Name == name {
			return i
		}
	}
	return nil
}

// FindColumn finds the column by it's name.
func (t *Table) FindColumn(name string) *Column {
	for _, c := range t.Columns {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (t *Table) autoMigrate(ctx context.Context, conn internal.Connection) error {
	if !HasTable(ctx, conn, t) {
		definitions, err := t.definition()
		if err != nil {
			return err
		}

		for _, def := range definitions {
			log.Debugf("AutoMigrate Query: \n%s", def)
			if _, err := conn.Exec(ctx, def); err != nil {
				return err
			}
		}
		return nil
	}
	// Iterate over columns otherwise
	for _, c := range t.Columns {
		if !HasColumn(ctx, conn, t, c) {
			dt, err := findDataType(c.Field())
			if err != nil {
				return err
			}

			if dtt, ok := dt.(ExternalDataTyper); ok {
				if _, err := conn.Exec(ctx, dtt.ExternalFunction(c.Field())); err != nil {
					return err
				}
			} else {
				query := fmt.Sprintf("ALTER TABLE %s.%s ADD %s %s;", quoteIdentifier(t.Schema), t.Name, c.Name, dt.GetName(c.Field()))
				log.Debugf("Updating column: %s, DB Query: \n%s", c.Name, query)
				if _, err := conn.Exec(ctx, query); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *Table) autoMigrateConstraints(ctx context.Context, conn internal.Connection) error {
	for _, column := range t.Columns {
		for _, constr := range column.Constraints {
			if !constr.DBChecker(ctx, conn, t, column) {
				def, err := constr.SQLName(t, column)
				if err != nil {
					return err
				}

				log.Debugf("AutoMigrate Query: \n%s", def)
				if _, err = conn.Exec(ctx, def); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *Table) createColumn(field *mapping.StructField, tags []*mapping.FieldTag) (err error) {
	// Insert the column
	c := &Column{field: field, Table: t}

	// Store the column on the 'ColumnKey' in the Store
	field.StoreSet(ColumnKey, c)
	for _, tag := range tags {
		// get the TagSetter function
		setter, ok := TagSetterFunctions[tag.Key]
		if !ok {
			log.Warningf("Model: '%s', Field: '%s' Struct Tag: '%s' not recognized", field.ModelStruct().Type().Name(), field.Name(), tag.Key)
			continue
		}

		if err = setter(field, tag); err != nil {
			return err
		}
	}
	// set the name if not set by tag setters
	if c.Name == "" {
		c.setName(field)
	}
	// set the type if not set by tag setters
	if c.Type == nil {
		c.Type, err = findDataType(field)
		if err != nil {
			return err
		}
	}
	c.setConstraints()
	t.Columns = append(t.Columns, c)

	return nil
}

func (t *Table) createIndex(name string, iType IndexType, c *Column) (i *Index) {
	if name != "" {
		i = t.FindIndex(name)
	}

	if i == nil {
		// if the name is not defined get new index name
		if name == "" {
			name = t.newIndexName(c)
		}

		i = &Index{Name: name, Type: iType}
	}

	i.Columns = append(i.Columns, c)
	return i
}

// Definition gets the Table sql definition
func (t *Table) definition() ([]string, error) {
	sb := &strings.Builder{}

	sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	sb.WriteString(quoteIdentifier(t.Schema))
	sb.WriteRune('.')
	sb.WriteString(quoteIdentifier(t.Name))
	sb.WriteString(" (\n")

	var inlineColumns int

	for _, c := range t.Columns {
		dt, err := findDataType(c.Field())
		if err != nil {
			return nil, err
		}

		if _, ok := dt.(ExternalDataTyper); !ok {
			inlineColumns++
		}
	}

	var i int

	for _, c := range t.Columns {
		dt, err := findDataType(c.Field())
		if err != nil {
			return nil, err
		}

		if _, ok := dt.(ExternalDataTyper); !ok {
			sb.WriteString(c.Name)
			sb.WriteString(" ")

			sb.WriteString(dt.GetName(c.Field()))
		} else {
			continue
		}

		// sb.WriteString(c.Type)

		if i < inlineColumns-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
		i++
	}
	sb.WriteString(");")

	var result = []string{sb.String()}

	// write external data types
	for _, c := range t.Columns {
		dt, err := findDataType(c.Field())
		if err != nil {
			return nil, err
		}
		if dtt, ok := dt.(ExternalDataTyper); ok {
			result = append(result, dtt.ExternalFunction(c.Field()))
		}
	}

	return result, nil
}

// newIndexName creates new index for the column.
// the index is of form 'schema_table_column_idx_1'
func (t *Table) newIndexName(column *Column) string {
	return fmt.Sprintf("%s_%s_%s_idx_%d", t.Schema, t.Name, column.Name, len(column.Indexes)+1)
}

// ModelsTable gets the model table
func ModelsTable(m *mapping.ModelStruct) (*Table, error) {
	return modelsTable(m)
}

// ModelsTableName gets the table name for the provided model
func ModelsTableName(m *mapping.ModelStruct) (string, error) {
	t, err := modelsTable(m)
	if err != nil {
		return "", err
	}
	return t.Name, nil
}

func modelsTable(m *mapping.ModelStruct) (*Table, error) {
	t, ok := m.StoreGet(TableKey)
	if !ok {
		return nil, errors.NewDetf(controller.ClassInternal, "model's: '%s' doesn't have table stored", m.Type().Name())
	}

	tb, ok := t.(*Table)
	if !ok {
		log.Errorf("The table stored within model's: '%s' Store is not a '*migrate.Table' - '%T'", m.Collection(), t)
		return nil, errors.NewDetf(controller.ClassInternal, "stored table for model: '%s' is not a *migrate.Table", m.Type().Name())
	}
	return tb, nil
}

// TableNamer is the interface that allows to get the table name for given model
type TableNamer interface {
	TableName() string
}

// SchemaNamer is the interface for the models that needs non default 'postgres' schema name.
type SchemaNamer interface {
	PQSchemaName() string
}
