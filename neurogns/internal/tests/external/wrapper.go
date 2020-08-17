package external

//go:generate neurogns models methods --format=goimports --single-file .

// Int is the testing integer wrapper.
type Int int

// Model is testing external model.
type Model struct {
	ID int
}
