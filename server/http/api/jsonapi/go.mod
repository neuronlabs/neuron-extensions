module github.com/neuronlabs/neuron-plugins/server/http/api/jsonapi

replace (
	github.com/neuronlabs/neuron => ./../../../../../neuron
	github.com/neuronlabs/neuron-plugins/codec/jsonapi => ./../../../../codec/jsonapi
	github.com/neuronlabs/neuron-plugins/server/http => ./../../
	github.com/neuronlabs/neuron/errors => ./../../../../../neuron/errors
)

go 1.12

require (
	github.com/julienschmidt/httprouter v1.3.0
	github.com/neuronlabs/neuron v0.15.0
	github.com/neuronlabs/neuron-plugins/codec/jsonapi v0.0.0
	github.com/neuronlabs/neuron-plugins/server/http v0.0.0-20200717092015-ffd984ac8f41
	github.com/neuronlabs/neuron/errors v1.1.1-0.20190801002318-9535ebe7d446
)
