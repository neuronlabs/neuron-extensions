package migrate

import (
	"context"
	"fmt"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/internal"
)

// field tag constraints
const (
	cNotNull = "notnull"
	cUnique  = "unique"
	cForeign = "foreign"
)

// Constraint defines the postgres constraint type.
type Constraint struct {
	Name      string
	SQLName   func(t *Table, c *Column) (string, error)
	DBChecker func(context.Context, internal.Connection, *Table, *Column) bool
}

func uniqueConstraintName(c *Column) string {
	return fmt.Sprintf("unique_%s_%s", c.Table.Name, c.Name)
}

var (
	// CNotNull is the not null constraint
	CNotNull = &Constraint{Name: cNotNull, SQLName: func(t *Table, c *Column) (string, error) {
		return fmt.Sprintf("ALTER TABLE %s.%s ALTER COLUMN %s SET NOT NULL;",
			quoteIdentifier(t.Schema), quoteIdentifier(t.Name),
			c.Name), nil
	},
		DBChecker: HasNotNullConstraint,
	}

	// CUnique is the 'unique' constraint.
	CUnique = &Constraint{Name: cUnique, SQLName: func(t *Table, c *Column) (string, error) {
		return fmt.Sprintf("ALTER TABLE %s.%s ADD CONSTRAINT %s UNIQUE (%s);",
			quoteIdentifier(t.Schema), quoteIdentifier(t.Name),
			uniqueConstraintName(c), c.Name,
		), nil
	},
		DBChecker: HasUniqueConstraint,
	}

	// CPrimaryKey is the Primary key constraint.
	CPrimaryKey = &Constraint{Name: "primary", SQLName: func(t *Table, c *Column) (string, error) {
		return fmt.Sprintf("ALTER TABLE %s.%s ADD PRIMARY KEY (%s);",
			quoteIdentifier(t.Schema), quoteIdentifier(t.Name), c.Name), nil
	},
		DBChecker: HasPrimaryKey,
	}

	// CForeignKey is the Foreign key constraint.
	CForeignKey = &Constraint{Name: "foreign", SQLName: func(t *Table, c *Column) (string, error) {
		relatedField := c.Field().Relationship().Struct().Primary()

		relatedTable, err := modelsTable(relatedField.ModelStruct())
		if err != nil {
			return "", err
		}

		relatedPrimaryColumn, err := fieldsColumn(relatedField)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("ALTER TABLE %s.%s ADD FOREIGN KEY (%s) REFERENCES %s.%s(%s);",
			quoteIdentifier(t.Schema), quoteIdentifier(t.Name), c.Name,
			quoteIdentifier(relatedTable.Schema), quoteIdentifier(relatedTable.Name), relatedPrimaryColumn.Name,
		), nil
	},
		DBChecker: HasForeignKey,
	}
)
