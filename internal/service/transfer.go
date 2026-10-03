package service

import (
	"context"

	"github.com/pravinkanna/wallet-transfer-assignment/internal/domain"
)

// Tx is the transaction-scoped repository the transfer workflow runs on.
type Tx interface{}

// Repository runs fn in one database transaction, passing it a
// transaction-scoped repository of type T.
type Repository[T Tx] interface {
	InTx(ctx context.Context, fn func(T) error) error
}

// TransferService runs the transfer workflow (design §5).
type TransferService struct{}

// New returns a TransferService that runs on repo. It is generic so the
// repository can pass its own transaction type to fn without importing this
// package (design §3).
func New[T Tx](repo Repository[T]) *TransferService {
	return &TransferService{}
}

// Transfer runs a validated request in one transaction and returns the
// transfer in its final state.
func (s *TransferService) Transfer(ctx context.Context, req domain.TransferRequest) (domain.Transfer, error) {
	return domain.Transfer{}, nil
}
