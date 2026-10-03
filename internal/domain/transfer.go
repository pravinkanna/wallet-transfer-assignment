package domain

import "errors"

// TransferState is where a transfer is in its lifecycle. PROCESSED and FAILED
// are final.
type TransferState string

const (
	StatePending   TransferState = "PENDING"
	StateProcessed TransferState = "PROCESSED"
	StateFailed    TransferState = "FAILED"
)

// ErrInvalidTransition means a transfer was asked to leave a final state.
var ErrInvalidTransition = errors.New("invalid transfer state transition")

// Transfer moves Amount from one wallet to another.
type Transfer struct {
	ID             string
	IdempotencyKey string
	FromWalletID   string
	ToWalletID     string
	Amount         int64
	State          TransferState
}

// NewTransfer creates a PENDING transfer for a validated request.
func NewTransfer(id string, req TransferRequest) Transfer {
	return Transfer{}
}

// MarkProcessed moves a PENDING transfer to PROCESSED.
func (t *Transfer) MarkProcessed() error {
	return nil
}

// MarkFailed moves a PENDING transfer to FAILED.
func (t *Transfer) MarkFailed() error {
	return nil
}
