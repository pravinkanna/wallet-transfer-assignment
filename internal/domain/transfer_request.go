package domain

// TransferRequest is a validated request to move Amount from one wallet to
// another.
type TransferRequest struct {
	IdempotencyKey string
	FromWalletID   string
	ToWalletID     string
	Amount         int64
}

// NewTransferRequest validates the request fields in the order of API spec §6
// steps 3–5. amount is the raw JSON text of the amount field. A missing or
// null field is passed as "".
func NewTransferRequest(idempotencyKey, fromWalletID, toWalletID, amount string) (TransferRequest, error) {
	return TransferRequest{}, nil
}
