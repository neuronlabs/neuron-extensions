package migrate

import (
	"reflect"
	"strings"

	"github.com/neuronlabs/strcase"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/log"
)

// Column is a postgres field kind.
type Column struct {
	// Name defines the column name
	Name string
	// Type is the column data type
	Type DataTyper
	// Variables defines the data type specific variables
	Variables []string
	// Constraints defines column constrains
	Constraints []*Constraint
	// Indexes defines the column indexes
	Indexes []*Index
	// Table is the pointer to the column's table
	Table *Table

	field *mapping.StructField
}

// Field is the column's related *mapping.StructField.
func (c *Column) Field() *mapping.StructField {
	return c.field
}

func (c *Column) isNotNull() bool {
	for _, cstr := range c.Constraints {
		if cstr == CNotNull {
			return true
		}
	}
	return false
}

func (c *Column) setName(field *mapping.StructField) {
	c.Name = strcase.ToSnake(field.Name())
}

func (c *Column) setConstraints() {
	// set the field's primary index
	if c.field.Kind() == mapping.KindPrimary {
		c.Constraints = append(c.Constraints, CPrimaryKey)
	}

	// check if nullable
	rf := c.field.ReflectField()

	if !c.isNotNull() {
		if rf.Type.Kind() != reflect.Ptr && !strings.Contains(strings.ToLower(rf.Type.Name()), "null") {
			// otherwise mark the column as it is not null
			c.Constraints = append(c.Constraints, CNotNull)
			c.Field().StoreSet(NotNullKey, struct{}{})
		}
	}
}

// ColumnCreator is the function that creates custom.
type ColumnCreator func(c *Column) string

// FieldColumnName gets the column name for the provided field.
func FieldColumnName(field *mapping.StructField) (string, error) {
	c, err := fieldsColumn(field)
	if err != nil {
		return "", err
	}
	return c.Name, nil
}

// FieldsColumn gets the column name for the provided field.
func FieldsColumn(field *mapping.StructField) (*Column, error) {
	c, err := fieldsColumn(field)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// FieldIndexes gets the column's indexes.
func FieldIndexes(field *mapping.StructField) ([]*Index, error) {
	c, err := fieldsColumn(field)
	if err != nil {
		return nil, err
	}

	return c.Indexes, nil
}

func fieldsColumn(field *mapping.StructField) (*Column, error) {
	col, ok := field.StoreGet(ColumnKey)
	if !ok {
		log.Debugf("No column found in the field: %s store.", field.NeuronName())
		return nil, errors.WrapDetf(errors.ErrInternal, "no column found in the field's '%s' store", field.Name())
	}

	// parse the column
	c, ok := col.(*Column)
	if !ok {
		log.Errorf("Column in the field's store is not a '*migrate.Column' : '%T'", col)
		return nil, errors.WrapDetf(errors.ErrInternal, "stored column for field: '%s' is not a *migrate.Column", field.Name())
	}

	return c, nil
}
