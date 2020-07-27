module github.com/neuronlabs/neuron-extensions/codec/json

replace (
    github.com/neuronlabs/neuron => ./../../../neuron
    github.com/neuronlabs/neuron/errors => ./../../../neuron/errors
)

go 1.14

require github.com/neuronlabs/neuron v0.0.0-00010101000000-000000000000
