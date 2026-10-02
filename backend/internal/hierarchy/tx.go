package hierarchy

import (
	"context"

	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	sqlcNode "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repos struct {
	Store     Store
	NodeStore NodeStore
}

type Transactor interface {
	InTx(ctx context.Context, fn func(Repos) error) error
}

type PgTransactor struct {
	pool *pgxpool.Pool
}

func NewPgTransactor(pool *pgxpool.Pool) *PgTransactor {
	return &PgTransactor{pool: pool}
}

func (t *PgTransactor) InTx(ctx context.Context, fn func(Repos) error) error {
	return pgx.BeginFunc(ctx, t.pool, func(tx pgx.Tx) error {
		return fn(Repos{
			Store:     sqlcHierarchy.New(tx),
			NodeStore: sqlcNode.New(tx),
		})
	})
}
