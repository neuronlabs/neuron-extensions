// +build integrate

package migrate

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/internal"
)

var cfg pgxpool.Config

// TestAutoMigrateModels tests the auto migration of the models
func TestAutoMigrateModels(t *testing.T) {
	repoCfg := internal.TestingConfig(t)
	models := []mapping.Model{&Model{}, &BasicModel{}}

	m := tCtrl(t, models...)

	ctx := context.Background()
	db, err := pgxpool.ConnectConfig(ctx, repoCfg)
	require.NoError(t, err)

	defer db.Close()

	for _, model := range models {
		modelStruct, ok := m.GetModelStruct(model)
		require.True(t, ok)

		err := AutoMigrateModel(ctx, db, modelStruct)
		require.NoError(t, err)

		table, err := modelsTable(modelStruct)
		if err != nil {
			continue
		}
		_, err = db.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s.%s;", quoteIdentifier(table.Schema), table.Name))
		if err != nil {
			log.Debugf("Error while dropping table: %v", err)
		}
	}
}
