// +build integrate

package postgres

import (
	"context"
	"testing"
	"time"

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

func TestDialAndClose(t *testing.T) {
	db := testingDB(t, true)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := db.Dial(ctx)
	require.NoError(t, err)

	err = db.Close(ctx)
	require.NoError(t, err)
}
