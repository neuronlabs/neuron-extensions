module github.com/neuronlabs/neuron-mocks

replace (
    github.com/neuronlabs/neuron => ./../../../neuron
    github.com/neuronlabs/neuron/errors => ./../../../neuron/errors
)
go 1.11

require (
	github.com/google/uuid v1.1.1
	github.com/neuronlabs/neuron v0.0.0
	github.com/stretchr/testify v1.4.0
	golang.org/x/text v0.3.2 // indirect
)
