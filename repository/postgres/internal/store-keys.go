package internal

var (
	// IncrementorKey is the scope's context key used to save current incrementor value.
	IncrementorKey = incrementorKey{}
)

type incrementorKey struct{}
