module github.com/neuronlabs/neuron-extensions/server/http

replace (
	github.com/neuronlabs/neuron => ./../../../neuron
)

require (
	github.com/julienschmidt/httprouter v1.3.0
	github.com/neuronlabs/brotli v1.0.1
	github.com/neuronlabs/neuron v0.0.0
)

go 1.12
