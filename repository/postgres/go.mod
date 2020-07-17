module github.com/neuronlabs/neuron-plugins/repository/postgres

go 1.11

replace (
	github.com/neuronlabs/neuron => ./../../../neuron
	github.com/neuronlabs/neuron/errors => ./../../../neuron/errors
)

require (
	github.com/google/uuid v1.1.1
	github.com/jackc/pgconn v1.5.0
	github.com/jackc/pgtype v1.3.0
	github.com/jackc/pgx/v4 v4.6.0
	github.com/neuronlabs/inflection v1.0.1
	github.com/neuronlabs/neuron v0.0.0-00010101000000-000000000000
	github.com/neuronlabs/neuron/errors v0.0.0-20200514135224-6481a6918f10
	github.com/neuronlabs/strcase v1.0.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shopspring/decimal v0.0.0-20190905144223-a36b5d85f337 // indirect
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37 // indirect
)
