package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/repository"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/log"
)

// compile time check for the repository.Transactioner interface.
var _ repository.Transactioner = &Postgres{}

// Begin starts a transaction for the given scope.
// Implements Begin method of the query.Transactioner interface.
func (p *Postgres) Begin(ctx context.Context, tx *query.Transaction) error {
	var isolation pgx.TxIsoLevel
	txOpts := pgx.TxOptions{IsoLevel: isolation}
	if tx.Options != nil {
		switch tx.Options.Isolation {
		case query.LevelDefault:
		case query.LevelSerializable:
			isolation = pgx.Serializable
		case query.LevelReadCommitted:
			isolation = pgx.ReadCommitted
		case query.LevelReadUncommitted:
			isolation = pgx.ReadUncommitted
		case query.LevelRepeatableRead, query.LevelSnapshot:
			isolation = pgx.RepeatableRead
		default:
			return errors.NewDetf(query.ClassTxState, "unsupported isolation level: %s", tx.Options.Isolation.String())
		}
		txOpts.IsoLevel = isolation
		if tx.Options.ReadOnly {
			txOpts.AccessMode = pgx.ReadOnly
		}
	}

	pgxTx, err := p.ConnPool.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}

	p.Transactions[tx.ID] = pgxTx
	return nil
}

// Commit commits the scope's transaction
// Implements Commit method from the query.Commiter interface
func (p *Postgres) Commit(ctx context.Context, tx *query.Transaction) error {
	if tx == nil {
		return errors.NewDet(query.ClassTxInvalid, "scope's transaction is nil")
	}

	pgxTx, ok := p.Transactions[tx.ID]
	if !ok {
		log.Errorf("Transaction: '%s' no mapped SQL transaction found", tx.ID)
		return errors.NewDet(query.ClassTxInvalid, "no mapped sql transaction found for the scope")
	}
	defer delete(p.Transactions, tx.ID)

	return pgxTx.Commit(ctx)
}

// Rollback rolls back the transaction for given scope
func (p *Postgres) Rollback(ctx context.Context, tx *query.Transaction) error {
	if tx == nil {
		return errors.NewDet(query.ClassTxInvalid, "scope's transaction is nil")
	}
	pgxTx, ok := p.Transactions[tx.ID]
	if !ok {
		log.Errorf("Transaction: '%s' no mapped SQL transaction found", tx.ID)
		return errors.NewDet(query.ClassTxInvalid, "no mapped sql transaction found for the scope")
	}
	defer delete(p.Transactions, tx.ID)

	return pgxTx.Rollback(ctx)
}
