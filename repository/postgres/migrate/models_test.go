package migrate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/mapping"
)

//go:generate neurogonesis models methods --format=goimports --type=Model,BasicModel --single-file .

type Model struct {
	ID         int        `neuron:"type=primary"`
	Attr       string     `neuron:"type=attr" db:"name=attribute"`
	SnakeCased string     `neuron:"type=attr"`
	CreatedAt  time.Time  `neuron:"type=attr"`
	UpdatedAt  time.Time  `neuron:"type=attr"`
	DeletedAt  *time.Time `neuron:"type=attr"`
}

type BasicModel struct {
	ID        int        `neuron:"type=primary"`
	String    string     `neuron:"type=attr" db:";unique"`
	Timed     time.Time  `neuron:"type=attr"`
	PtrTime   *time.Time `neuron:"type=attr"`
	Int       int        `neuron:"type=attr" db:";notnull"`
	Int16     int16      `neuron:"type=attr"`
	Varchar20 string     `neuron:"type=attr" db:"type=varchar(20)"`
	Float32   float32    `neuron:"type=attr"`
	IntArray  [3]int
	IntSlice  []int
}

// TestParseModel tests the extraction of the pq tags
func TestParseModel(t *testing.T) {
	t.Run("WithTimeFields", func(t *testing.T) {
		// type the some model
		some := &Model{}
		m := testingModelMap(t, some)

		mStruct, err := m.ModelStruct(some)
		require.NoError(t, err)

		for _, field := range mStruct.StructFields() {
			switch field.Name() {
			case "ID":
				assert.Equal(t, "id", field.DatabaseName)
			case "Attr":
				assert.Equal(t, "attribute", field.DatabaseName)
			case "SnakeCased":
				assert.Equal(t, "snake_cased", field.DatabaseName)
			}
		}
	})
}

func testingModelMap(t *testing.T, models ...mapping.Model) *mapping.ModelMap {
	t.Helper()

	m := mapping.New(
		mapping.WithNamingConvention(mapping.SnakeCase),
		mapping.WithDefaultDatabaseSchema("public"),
	)
	require.NoError(t, m.RegisterModels(models...))
	return m
}
