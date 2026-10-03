package handler

import (
	"net/http"

	"github.com/pravinkanna/wallet-transfer-assignment/internal/service"
)

// New returns the HTTP handler for the API (spec §1).
func New(svc *service.TransferService) http.Handler {
	return http.NewServeMux()
}
