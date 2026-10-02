package hierarchy

import (
	"context"

	"github.com/KubantsevAS/notree/backend/internal/db"
	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	"github.com/jackc/pgx/v5"
)

type Transactor interface {
	InTx(ctx context.Context, fn func(Store) error) error
}

type storeTransactor struct {
	tx db.Transactor
}

func newStoreTransactor(tx db.Transactor) *storeTransactor {
	return &storeTransactor{tx: tx}
}

func (t *storeTransactor) InTx(ctx context.Context, fn func(Store) error) error {
	return t.tx.InTx(ctx, func(tx pgx.Tx) error {
		return fn(sqlcHierarchy.New(tx))
	})
}
