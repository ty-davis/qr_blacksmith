package handlers

import (
	"database/sql"
	"net/http"

	"github.com/qr-blacksmith/backend/internal/auth"
	dbpkg "github.com/qr-blacksmith/backend/internal/db"
)

type ScanHandler struct {
	db *sql.DB
}

func NewScanHandler(database *sql.DB) *ScanHandler {
	return &ScanHandler{db: database}
}

func (h *ScanHandler) GetOverviewStats(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	stats, err := dbpkg.GetOverviewStats(h.db, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
