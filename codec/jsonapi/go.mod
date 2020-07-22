module github.com/neuronlabs/jsonapi

go 1.11

replace (
	github.com/neuronlabs/neuron => ./../../../neuron
	github.com/neuronlabs/neuron-plugins/server => ../../server
	github.com/neuronlabs/neuron/errors => ./../../../neuron/errors
)

require (
	github.com/neuronlabs/neuron v0.15.0
	github.com/neuronlabs/neuron-plugins/server/http v0.0.0-20200717092015-ffd984ac8f41
	github.com/neuronlabs/neuron/errors v1.1.1-0.20190801002318-9535ebe7d446
	github.com/stretchr/testify v1.4.0
	golang.org/x/tools v0.0.0-20200721032237-77f530d86f9a
)
