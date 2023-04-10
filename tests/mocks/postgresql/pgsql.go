package mocks

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type PgxPoolMock interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
