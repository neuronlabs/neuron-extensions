package input

import (
	"sort"
	"strings"

	"github.com/neuronlabs/strcase"
)

// MultiModel is a input for the multi model templates.
type MultiModel struct {
	PackageName string
	Imports     Imports
	Models      []*Model
}

// Model is a structure used to insert model into template.
type Model struct {
	TestFile                                                         bool
	FileName                                                         string
	PackageName                                                      string
	PackagePath                                                      string
	Imports                                                          Imports
	Name                                                             string
	Receiver                                                         string
	CollectionName                                                   string
	RepositoryName                                                   string
	Primary                                                          *Field
	Fields                                                           []*Field
	Fielder, SingleRelationer, MultiRelationer, CustomCollectionName bool
	Relations                                                        []*Field
	StructFields                                                     []*Field
	Receivers                                                        map[string]int
}

// AddImport adds 'imp' imported package if it doesn't exists already.
func (m *Model) AddImport(imp string) {
	m.Imports.Add(imp)
}

// Collection returns model's collection.
func (m *Model) Collection() *Collection {
	camelCollection := strcase.ToCamel(m.CollectionName)
	return &Collection{
		Name:         "NRN_" + camelCollection,
		VariableName: "NRN_" + camelCollection,
		QueryBuilder: strcase.ToLowerCamel(m.CollectionName + "QueryBuilder"),
	}
}

// CollectionInput returns template collection input for given model.
func (m *Model) CollectionInput(packageName string, isModelImported bool) *CollectionInput {
	c := &CollectionInput{
		PackageName: packageName,
		Imports: []string{
			"context",
			"github.com/neuronlabs/neuron/core",
			"github.com/neuronlabs/neuron/database",
			"github.com/neuronlabs/neuron/errors",
			"github.com/neuronlabs/neuron/mapping",
			"github.com/neuronlabs/neuron/query",
			"github.com/neuronlabs/neuron/query/filter",
		},
		Model:         m,
		Collection:    m.Collection(),
		ModelImported: true,
	}
	if c.PackageName == "" {
		c.PackageName = m.PackageName
	}
	if isModelImported {
		c.Imports.Add(m.PackagePath)
	}

	for _, relation := range m.Relations {
		if relation.IsImported {
			for _, mi := range m.Imports {
				if strings.HasSuffix(mi, relation.Selector) {
					c.Imports.Add(mi)
					break
				}
			}
		}
	}
	return c
}

// SortFields sorts the fields in the model.
func (m *Model) SortFields() {
	sort.Slice(m.Fields, func(i, j int) bool {
		return m.Fields[i].Index < m.Fields[j].Index
	})
}
