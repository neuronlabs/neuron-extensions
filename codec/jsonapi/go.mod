module github.com/neuronlabs/jsonapi

go 1.11

replace (
	github.com/neuronlabs/neuron => ./../../../neuron
	github.com/neuronlabs/neuron/errors => ./../../../neuron/errors
)

require (
	github.com/neuronlabs/neuron v0.15.0
	github.com/neuronlabs/neuron-mocks v0.14.2
	github.com/neuronlabs/neuron/errors v1.1.1-0.20190801002318-9535ebe7d446
	github.com/stretchr/testify v1.4.0
)
