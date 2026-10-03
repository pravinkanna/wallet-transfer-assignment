package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pravinkanna/wallet-transfer-assignment/internal/domain"
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
type Tx struct {
	tx pgx.Tx
}

// InTx runs fn in one transaction. It commits if fn returns nil and rolls
// back otherwise.
func (r *Repository) InTx(ctx context.Context, fn func(*Tx) error) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&Tx{tx: tx})
	})
}

// InsertTransfer stores a new transfer as PENDING.
func (t *Tx) InsertTransfer(ctx context.Context, transfer domain.Transfer) error {
	_, err := t.tx.Exec(ctx, `
		INSERT INTO transfers (id, idempotency_key, from_wallet_id, to_wallet_id, amount, state)
		VALUES ($1, $2, $3, $4, $5, 'PENDING')`,
		transfer.ID, transfer.IdempotencyKey, transfer.FromWalletID, transfer.ToWalletID, transfer.Amount)
	if err != nil {
		return fmt.Errorf("insert transfer: %w", err)
	}
	return nil
}

// DebitWallet subtracts amount from a wallet's balance.
func (t *Tx) DebitWallet(ctx context.Context, walletID string, amount int64) error {
	_, err := t.tx.Exec(ctx, `UPDATE wallets SET balance = balance - $2 WHERE id = $1`, walletID, amount)
	if err != nil {
		return fmt.Errorf("debit wallet %s: %w", walletID, err)
	}
	return nil
}

// CreditWallet adds amount to a wallet's balance.
func (t *Tx) CreditWallet(ctx context.Context, walletID string, amount int64) error {
	_, err := t.tx.Exec(ctx, `UPDATE wallets SET balance = balance + $2 WHERE id = $1`, walletID, amount)
	if err != nil {
		return fmt.Errorf("credit wallet %s: %w", walletID, err)
	}
	return nil
}

// InsertLedgerEntry stores one ledger entry.
func (t *Tx) InsertLedgerEntry(ctx context.Context, entry domain.LedgerEntry) error {
	_, err := t.tx.Exec(ctx, `
		INSERT INTO ledger_entries (transfer_id, wallet_id, type, amount)
		VALUES ($1, $2, $3, $4)`,
		entry.TransferID, entry.WalletID, entry.Type, entry.Amount)
	if err != nil {
		return fmt.Errorf("insert ledger entry: %w", err)
	}
	return nil
}

// UpdateTransferState moves a PENDING transfer to state. Like the domain, it
// refuses to move a transfer that is no longer PENDING.
func (t *Tx) UpdateTransferState(ctx context.Context, transferID string, state domain.TransferState) error {
	tag, err := t.tx.Exec(ctx, `
		UPDATE transfers SET state = $2 WHERE id = $1 AND state = 'PENDING'`,
		transferID, state)
	if err != nil {
		return fmt.Errorf("update transfer state: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("%w: transfer %s is not PENDING", domain.ErrInvalidTransition, transferID)
	}
	return nil
}
