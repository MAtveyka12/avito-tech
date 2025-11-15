package tx

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Raw() pgx.Tx
}

type transaction struct {
	tx pgx.Tx
}

func NewTransaction(tx pgx.Tx) Transaction {
	return &transaction{tx: tx}
}

func (t *transaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *transaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (t *transaction) Raw() pgx.Tx {
	return t.tx
}

