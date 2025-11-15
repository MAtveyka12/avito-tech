package tx

import (
	"context"
	"log/slog"
)

type TxManager interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}

type txManager struct {
	ctxManager CtxManager
	logger     *slog.Logger
}

func NewTxManager(ctxManager CtxManager, logger *slog.Logger) TxManager {
	return &txManager{
		ctxManager: ctxManager,
		logger:     logger,
	}
}

func (m *txManager) Do(ctx context.Context, fn func(context.Context) error) error {
	m.logger.Debug("starting transaction")

	tx, err := m.ctxManager.Default(ctx)
	if err != nil {
		m.logger.Error("failed to start transaction", slog.Any("error", err))
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txCtx := m.ctxManager.WithTx(ctx, tx)

	if err := fn(txCtx); err != nil {
		m.logger.Warn("transaction rollback", slog.Any("error", err))
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		m.logger.Error("failed to commit transaction", slog.Any("error", err))
		return err
	}

	m.logger.Debug("transaction committed successfully")

	return nil
}

