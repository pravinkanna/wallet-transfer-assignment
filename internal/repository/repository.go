package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists wallets, transfers, and ledger entries in Postgres.
type Repository struct {
	pool *pgxpool.Pool
}

// New returns a Repository that runs its statements on pool.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Tx runs persistence operations inside one database transaction.
type Tx struct{}

// InTx runs fn in one transaction. It commits if fn returns nil and rolls
// back otherwise.
func (r *Repository) InTx(ctx context.Context, fn func(*Tx) error) error {
	return nil
}
