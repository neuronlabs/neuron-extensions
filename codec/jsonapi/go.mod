module github.com/neuronlabs/neuron-extensions/codec/jsonapi

replace (
	github.com/neuronlabs/neuron => ./../../../neuron
)

go 1.12

require (
	github.com/neuronlabs/neuron v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.6.1
)
