module github.com/neuronlabs/neuron-extensions/codec/jsonapi

replace (
	github.com/neuronlabs/neuron => ./../../../neuron
	github.com/neuronlabs/neuron/errors => ./../../../neuron/errors
)

go 1.12

require (
	github.com/neuronlabs/neuron v0.0.0-00010101000000-000000000000
	github.com/neuronlabs/neuron/errors v0.0.0-20200511120829-fff1f8cf09c7
	github.com/stretchr/testify v1.6.1
)
