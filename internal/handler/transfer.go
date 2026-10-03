package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pravinkanna/wallet-transfer-assignment/internal/domain"
	"github.com/pravinkanna/wallet-transfer-assignment/internal/service"
)

// New returns the HTTP handler for the API (spec §1).
func New(svc *service.TransferService) http.Handler {
	h := &transferHandler{svc: svc}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /transfers", h.createTransfer)
	return mux
}

type transferHandler struct {
	svc *service.TransferService
}

// transferResponse is the transfer body from spec §3.
type transferResponse struct {
	TransferID string               `json:"transferId"`
	State      domain.TransferState `json:"state"`
}

func (h *transferHandler) createTransfer(w http.ResponseWriter, r *http.Request) {
	req, err := decodeRequest(r.Body)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	transfer, err := h.svc.Transfer(r.Context(), req)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, transferResponse{TransferID: transfer.ID, State: transfer.State})
}

// decodeRequest turns a request body into a validated domain request
// (design §5).
func decodeRequest(body io.Reader) (domain.TransferRequest, error) {
	var fields map[string]json.RawMessage
	if err := json.NewDecoder(body).Decode(&fields); err != nil {
		return domain.TransferRequest{}, err
	}
	key, err := stringField(fields, "idempotencyKey")
	if err != nil {
		return domain.TransferRequest{}, err
	}
	from, err := stringField(fields, "fromWalletId")
	if err != nil {
		return domain.TransferRequest{}, err
	}
	to, err := stringField(fields, "toWalletId")
	if err != nil {
		return domain.TransferRequest{}, err
	}
	return domain.NewTransferRequest(key, from, to, string(fields["amount"]))
}

// stringField unquotes a JSON string field. A missing field or null gives "".
func stringField(fields map[string]json.RawMessage, name string) (string, error) {
	var value string
	if raw, ok := fields[name]; ok {
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", err
		}
	}
	return value, nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
