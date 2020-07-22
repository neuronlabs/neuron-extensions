module github.com/neuronlabs/neuron-plugins/server/http

go 1.11

replace (
	github.com/neuronlabs/neuron => ./../../../neuron
	github.com/neuronlabs/neuron/errors => ./../../../neuron/errors
)

require (
	github.com/julienschmidt/httprouter v1.3.0
	github.com/neuronlabs/brotli v1.0.1
	github.com/neuronlabs/jsonapi-handler v0.0.1
	github.com/neuronlabs/neuron v0.15.0
	github.com/neuronlabs/neuron/errors v0.0.0-20200511120829-fff1f8cf09c7
)
