module github.com/neuronlabs/neuron-extensions/auth/account

replace github.com/neuronlabs/neuron => ./../../../neuron

go 1.12

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/neuronlabs/neuron v0.0.0-00010101000000-000000000000
	github.com/neuronlabs/neuron/errors v0.0.0-20200511120829-fff1f8cf09c7
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
)
