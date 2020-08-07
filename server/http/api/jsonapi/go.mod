module github.com/neuronlabs/neuron-extensions/server/http/api/jsonapi

replace (
	github.com/neuronlabs/neuron => ./../../../../../neuron
	github.com/neuronlabs/neuron-extensions/codec/jsonapi => ./../../../../codec/jsonapi
	github.com/neuronlabs/neuron-extensions/server/http => ./../../
)

go 1.12

require (
	github.com/julienschmidt/httprouter v1.3.0
	github.com/neuronlabs/neuron v0.15.0
	github.com/neuronlabs/neuron-extensions/codec/jsonapi v0.0.0
	github.com/neuronlabs/neuron-extensions/server/http v0.0.0-20200717092015-ffd984ac8f41
)
