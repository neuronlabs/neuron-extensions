package ast

import (
	"github.com/neuronlabs/neuron-extensions/neurogonesis/input"
)

func (g *ModelGenerator) checkModelMethods(m *input.Model) error {
	typeName := m.PackageName + "." + m.Name
	for _, method := range g.typeMethods[typeName] {
		if method.IsNeuronCollectionName() {
			m.CustomCollectionName = true
			m.CollectionName = method.ReturnStatement
		}
	}
	return nil
}
