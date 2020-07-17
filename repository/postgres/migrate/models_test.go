package migrate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/mapping"
)

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
	String    string     `neuron:"type=attr" db:"unique"`
	Timed     time.Time  `neuron:"type=attr"`
	PtrTime   *time.Time `neuron:"type=attr"`
	Int       int        `neuron:"type=attr" db:"notnull"`
	Int16     int16      `neuron:"type=attr"`
	Varchar20 string     `neuron:"type=attr" db:"type=varchar(20)"`
	Float32   float32    `neuron:"type=attr"`
}

//go:generate neuron-generator models methods --type=Model,BasicModel --single-file .

// TestParseModel tests the extraction of the pq tags
func TestParseModel(t *testing.T) {
	t.Run("WithTimeFields", func(t *testing.T) {
		// type the some model
		some := &Model{}
		m := tCtrl(t, some)

		mStruct, ok := m.GetModelStruct(some)
		require.True(t, ok)

		for _, field := range mStruct.StructFields() {
			cl, ok := field.StoreGet(ColumnKey)

			switch field.Name() {
			case "ID":
				if assert.True(t, ok) {
					col, ok := cl.(*Column)
					if assert.True(t, ok) {
						assert.Equal(t, "id", col.Name)
					}
				}
			case "Attr":
				if assert.True(t, ok) {
					col, ok := cl.(*Column)
					if assert.True(t, ok) {
						assert.Equal(t, "attribute", col.Name)
					}
				}
			case "SnakeCased":
				if assert.True(t, ok) {
					col, ok := cl.(*Column)
					if assert.True(t, ok) {
						assert.Equal(t, "snake_cased", col.Name)
					}
				}
			case "Nested":
			}
		}
	})
}

func tCtrl(t *testing.T, models ...mapping.Model) *mapping.ModelMap {
	t.Helper()

	m := mapping.NewModelMap(mapping.SnakeCase)
	require.NoError(t, m.RegisterModels(models...))
	require.NoError(t, prepareModels(m.Models()...))

	return m
}
