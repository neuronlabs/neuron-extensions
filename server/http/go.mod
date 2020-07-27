module github.com/neuronlabs/neuron-extensions/server/http

replace (
	github.com/neuronlabs/neuron => ./../../../neuron
    github.com/neuronlabs/neuron/errors => ./../../../neuron/errors
)

require (
	github.com/julienschmidt/httprouter v1.3.0
	github.com/neuronlabs/brotli v1.0.1
	github.com/neuronlabs/jsonapi-handler v0.0.1
	github.com/neuronlabs/neuron v0.0.0
	github.com/neuronlabs/neuron/errors v0.0.0-20200602214436-03d1e5cb9e8f
)

go 1.12
