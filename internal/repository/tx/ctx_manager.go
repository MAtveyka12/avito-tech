package tx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ctxKey struct{}

var txKey = ctxKey{}

type CtxManager interface {
	FromContext(ctx context.Context) (Transaction, error)
	WithTx(ctx context.Context, tx Transaction) context.Context
	Default(ctx context.Context) (Transaction, error)
}

type ctxManager struct {
	pool *pgxpool.Pool
}

func NewCtxManager(pool *pgxpool.Pool) CtxManager {
	return &ctxManager{pool: pool}
}

func (m *ctxManager) FromContext(ctx context.Context) (Transaction, error) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("no transaction found in context")
	}

	return NewTransaction(tx), nil
}

func (m *ctxManager) WithTx(ctx context.Context, tx Transaction) context.Context {
	return context.WithValue(ctx, txKey, tx.Raw())
}

func (m *ctxManager) Default(ctx context.Context) (Transaction, error) {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %s", err.Error())
	}

	return NewTransaction(tx), nil
}

