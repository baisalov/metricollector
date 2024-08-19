package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type ctxTxKey struct{}

type TransactionManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (m TransactionManager) Do(ctx context.Context, fn func(context.Context) error) (err error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if r := tx.Rollback(); r != nil {
				err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", r))
			}
		}
	}()

	ctx = context.WithValue(ctx, ctxTxKey{}, tx)

	err = fn(ctx)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
