package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/pravinkanna/wallet-transfer-assignment/internal/domain"
	"github.com/pravinkanna/wallet-transfer-assignment/internal/service"
)

// Errors for spec §6 steps 1–2, which only the handler can check.
var (
	errInvalidJSON  = errors.New("body must be a single JSON object")
	errUnknownField = errors.New("unknown field")
)

// errorCodes maps each client error to its status and code from spec §4.
var errorCodes = []struct {
	err    error
	status int
	code   string
}{
	{errInvalidJSON, http.StatusBadRequest, "INVALID_JSON"},
	{errUnknownField, http.StatusBadRequest, "UNKNOWN_FIELD"},
	{domain.ErrInvalidField, http.StatusBadRequest, "INVALID_FIELD"},
	{domain.ErrInvalidAmount, http.StatusBadRequest, "INVALID_AMOUNT"},
	{domain.ErrSameWallet, http.StatusBadRequest, "SAME_WALLET"},
	{domain.ErrWalletNotFound, http.StatusBadRequest, "WALLET_NOT_FOUND"},
	{domain.ErrIdempotencyKeyReused, http.StatusConflict, "IDEMPOTENCY_KEY_REUSED"},
}

// knownFields are the request fields from spec §2.
var knownFields = map[string]bool{
	"idempotencyKey": true,
	"fromWalletId":   true,
	"toWalletId":     true,
	"amount":         true,
}

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

// errorResponse is the error body from spec §4.
type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (h *transferHandler) createTransfer(w http.ResponseWriter, r *http.Request) {
	req, err := decodeRequest(r.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	transfer, err := h.svc.Transfer(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	status := http.StatusCreated
	if transfer.State == domain.StateFailed {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, transferResponse{TransferID: transfer.ID, State: transfer.State})
}

// decodeRequest turns a request body into a validated domain request
// (design §5).
func decodeRequest(body io.Reader) (domain.TransferRequest, error) {
	dec := json.NewDecoder(body)
	var fields map[string]json.RawMessage
	if err := dec.Decode(&fields); err != nil || fields == nil {
		return domain.TransferRequest{}, errInvalidJSON
	}
	if err := dec.Decode(&json.RawMessage{}); !errors.Is(err, io.EOF) {
		return domain.TransferRequest{}, errInvalidJSON
	}
	for name := range fields {
		if !knownFields[name] {
			return domain.TransferRequest{}, fmt.Errorf("%w: %s", errUnknownField, name)
		}
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
	amount := string(fields["amount"])
	if amount == "null" {
		amount = ""
	}
	return domain.NewTransferRequest(key, from, to, amount)
}

// stringField unquotes a JSON string field. A missing field or null gives "".
func stringField(fields map[string]json.RawMessage, name string) (string, error) {
	var value string
	if raw, ok := fields[name]; ok {
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", fmt.Errorf("%w: %s must be a string", domain.ErrInvalidField, name)
		}
	}
	return value, nil
}

// writeError writes the spec §4 body for err. Any error that is not a client
// error is a server error, and its details stay out of the response.
func writeError(w http.ResponseWriter, err error) {
	for _, e := range errorCodes {
		if errors.Is(err, e.err) {
			writeJSON(w, e.status, errorResponse{Error: errorDetail{Code: e.code, Message: err.Error()}})
			return
		}
	}
	writeJSON(w, http.StatusInternalServerError, errorResponse{Error: errorDetail{
		Code:    "INTERNAL_ERROR",
		Message: "internal server error",
	}})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
