// +build integrate

package postgres

import (
	"context"
	"testing"

	"github.com/neuronlabs/neuron/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
	p := testingRepository(testingDB(t, true))

	result, err := p.HealthCheck(context.Background())
	require.NoError(t, err)

	assert.Equal(t, repository.StatusPass, result.Status)
}
