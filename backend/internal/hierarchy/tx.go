package hierarchy

import (
	"context"

	"github.com/KubantsevAS/notree/backend/internal/db"
	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	sqlcNode "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/jackc/pgx/v5"
)

type Repos struct {
	Store     Store
	NodeStore NodeStore
}

type Transactor interface {
	InTx(ctx context.Context, fn func(Repos) error) error
}

type reposTransactor struct {
	tx db.Transactor
}

func newReposTransactor(tx db.Transactor) *reposTransactor {
	return &reposTransactor{tx: tx}
}

func (t *reposTransactor) InTx(ctx context.Context, fn func(Repos) error) error {
	return t.tx.InTx(ctx, func(tx pgx.Tx) error {
		return fn(Repos{
			Store:     sqlcHierarchy.New(tx),
			NodeStore: sqlcNode.New(tx),
		})
	})
}
