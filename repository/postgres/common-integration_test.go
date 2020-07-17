// +build integrate

package postgres

// import (
// 	"context"
// 	"testing"
// 	"time"
//
// 	"github.com/jackc/pgx"
// 	"github.com/spf13/viper"
// 	"github.com/stretchr/testify/require"
//
// 	"github.com/neuronlabs/neuron/config"
// 	"github.com/neuronlabs/neuron/controller"
// )
//
// var (
// 	tn                 = time.Now()
// 	testModelInstances = []*testModel{
// 		{ID: 1, AttrString: "some", Int: 2, CreatedAt: tn},
// 		{ID: 2, AttrString: "some", Int: 5, CreatedAt: tn},
// 		{ID: 3, AttrString: "some", Int: 1, CreatedAt: tn},
// 		{ID: 4, AttrString: "some", Int: 2, CreatedAt: tn},
// 	}
// )
//
// func prepareIntegrateRepository(t testing.TB) (*controller.Controller, *pgx.ConnPool) {
// 	t.Helper()
//
// 	v := viper.New()
//
// 	v.SetConfigName("test-config")
// 	v.AddConfigPath("internal")
//
// 	err := v.ReadInConfig()
// 	require.NoError(t, err)
//
// 	// set default config.
// 	cfg := config.Default()
// 	err = v.Unmarshal(cfg)
// 	require.NoError(t, err)
// 	// create new controller
// 	c, err := controller.New(cfg)
// 	require.NoError(t, err)
//
// 	ctx := context.Background()
// 	require.NoError(t, c.DialAll(ctx))
//
// 	// register models
// 	require.NoError(t, c.RegisterModels(&testModel{}))
// 	require.NoError(t, c.MigrateModels(ctx, &testModel{}))
//
// 	repo, err := c.GetRepository(testModel{})
// 	require.NoError(t, err)
//
// 	pg, ok := repo.(*Postgres)
// 	require.True(t, ok)
//
// 	// fill the test model
// 	fillTestModelTable(t, pg.ConnPool)
//
// 	return c, pg.ConnPool
// }
//
// func fillTestModelTable(t testing.TB, db *pgx.ConnPool) {
// 	t.Helper()
//
// 	strPtr := "value"
// 	testModelInstances[0].StringPtr = &strPtr
//
// 	for _, model := range testModelInstances {
// 		_, err := db.Exec(`INSERT INTO test_models (attr_string, string_ptr, int, created_at) VALUES($1,$2,$3,$4);`, model.AttrString, model.StringPtr, model.Int, model.CreatedAt)
// 		require.NoError(t, err)
// 	}
//
// }
//
// func deleteTestModelTable(t testing.TB, db *pgx.ConnPool) {
// 	t.Helper()
//
// 	_, err := db.Exec("DROP TABLE test_models;")
// 	require.NoError(t, err)
// }
