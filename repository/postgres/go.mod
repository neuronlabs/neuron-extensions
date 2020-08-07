module github.com/neuronlabs/neuron-extensions/repository/postgres

go 1.11

replace (
	github.com/neuronlabs/neuron => ./../../../neuron
)

require (
	github.com/google/uuid v1.1.1
	github.com/jackc/pgconn v1.6.4
	github.com/jackc/pgtype v1.4.2
	github.com/jackc/pgx/v4 v4.8.1
	github.com/neuronlabs/inflection v1.0.1
	github.com/neuronlabs/neuron v0.0.0-00010101000000-000000000000
	github.com/neuronlabs/strcase v1.0.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.5.1
)
