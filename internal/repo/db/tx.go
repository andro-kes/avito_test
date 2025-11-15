package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Tx interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context, q Querier) error) error
}

type tx struct {
	Pool *pgxpool.Pool
}

func NewTx(pool *pgxpool.Pool) Tx {
	return &tx{
		Pool: pool,
	}
}

func (t *tx) RunInTx(ctx context.Context, fn func(ctx context.Context, q Querier) error) error {
	px, err := t.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer px.Rollback(ctx)

	if fn(ctx, px) != nil {
		px.Rollback(ctx)
		return err
	}

	return px.Commit(ctx)
}