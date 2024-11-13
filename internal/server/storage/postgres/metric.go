package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
)

type MetricStorage struct {
	db *sql.DB
}

func NewMetricStorage(db *sql.DB) (*MetricStorage, error) {
	m := &MetricStorage{db: db}

	if err := m.migrate(); err != nil {
		return nil, err
	}

	return m, nil
}

func (s MetricStorage) All(ctx context.Context) (metrics []metric.Metric, err error) {
	query := `SELECT "type", "id", "delta", "value" FROM metrics WHERE true`

	var rows *sql.Rows

	if tx, ok := ctx.Value(ctxTxKey{}).(*sql.Tx); ok {
		rows, err = tx.QueryContext(ctx, query)
	} else {
		rows, err = s.db.QueryContext(ctx, query)
	}

	if err != nil {
		return nil, err
	}

	defer func() {
		if r := rows.Close(); r != nil {
			err = errors.Join(err, r)
		}
	}()

	for rows.Next() {
		var r rowMetric
		err = rows.Scan(&r.MType, &r.ID, &r.Delta, &r.Value)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, r.metric())
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (s MetricStorage) Get(ctx context.Context, t metric.Type, id string) (m metric.Metric, err error) {

	query := `SELECT "type", "id", "delta", "value" FROM metrics WHERE "type" = $1 AND "id" = $2`

	var stmt *sql.Stmt

	err = retry(func() error {
		stmt, err = s.db.PrepareContext(ctx, query)
		return err
	})
	if err != nil {
		return metric.Metric{}, fmt.Errorf("failed to make statmant: %w", err)
	}

	defer func() {
		if r := stmt.Close(); r != nil {
			err = errors.Join(err, r)
		}
	}()

	if tx, ok := ctx.Value(ctxTxKey{}).(*sql.Tx); ok {
		stmt = tx.StmtContext(ctx, stmt)
	}

	row := stmt.QueryRowContext(ctx, t, id)

	var r rowMetric

	err = retry(func() error {
		return row.Scan(&r.MType, &r.ID, &r.Delta, &r.Value)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return metric.Metric{}, metric.ErrMetricNotFound
		}
		return metric.Metric{}, err
	}

	return r.metric(), nil
}

func (s MetricStorage) Save(ctx context.Context, m metric.Metric) error {
	query := `INSERT INTO metrics ("type", "id", "delta", "value") VALUES ($1, $2, $3, $4)
		ON CONFLICT ("type", "id") DO UPDATE SET "delta"="excluded"."delta", "value"="excluded"."value"`

	var (
		stmt *sql.Stmt
		err  error
	)

	err = retry(func() error {
		stmt, err = s.db.PrepareContext(ctx, query)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to make statmant: %w", err)
	}

	defer func() {
		if r := stmt.Close(); r != nil {
			err = errors.Join(err, r)
		}
	}()

	if tx, ok := ctx.Value(ctxTxKey{}).(*sql.Tx); ok {
		stmt = tx.StmtContext(ctx, stmt)
	}

	err = retry(func() error {
		_, err = stmt.ExecContext(ctx, &m.MType, &m.ID, m.Delta, m.Value)
		return err
	})

	return err
}

type rowMetric struct {
	ID    string
	MType string
	Delta sql.NullInt64
	Value sql.NullFloat64
}

func (r rowMetric) metric() metric.Metric {
	var m metric.Metric

	m.ID = r.ID
	m.MType = metric.ParseType(r.MType)

	if r.Delta.Valid {
		m.Delta = &r.Delta.Int64
	}

	if r.Value.Valid {
		m.Value = &r.Value.Float64
	}

	return m
}

func (s MetricStorage) migrate() error {
	shame := `CREATE TABLE IF NOT EXISTS metrics (
    "type" VARCHAR(30) NOT NULL,
    "id" VARCHAR(30) NOT NULL,
    "delta" BIGINT,
    "value" DOUBLE PRECISION,
    PRIMARY KEY ("type", "id")
	);`

	err := retry(func() error {
		_, err := s.db.Exec(shame)
		return err
	})

	return err
}
