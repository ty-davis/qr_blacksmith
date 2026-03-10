package handlers

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/qr-blacksmith/backend/internal/auth"
	"github.com/qr-blacksmith/backend/internal/cache"
	dbpkg "github.com/qr-blacksmith/backend/internal/db"
	"github.com/qr-blacksmith/backend/internal/models"
	"github.com/qr-blacksmith/backend/internal/qr"
)

type QRCodeHandler struct {
	db      *sql.DB
	cache   *cache.RedirectCache
	baseURL string
}

func NewQRCodeHandler(database *sql.DB, c *cache.RedirectCache, baseURL string) *QRCodeHandler {
	return &QRCodeHandler{db: database, cache: c, baseURL: baseURL}
}

// effectiveBaseURL resolves the priority chain:
// batch base_url → user base_url → server BASE_URL env.
func effectiveBaseURL(batchBaseURL, userBaseURL *string, serverBaseURL string) string {
	if batchBaseURL != nil && *batchBaseURL != "" {
		return *batchBaseURL
	}
	if userBaseURL != nil && *userBaseURL != "" {
		return *userBaseURL
	}
	return serverBaseURL
}

func (h *QRCodeHandler) ListQRCodes(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	batchID := chi.URLParam(r, "id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}

	codes, total, err := dbpkg.ListQRCodes(h.db, batchID, userID, page, perPage)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if codes == nil {
		codes = []models.QRCode{}
	}
	for i := range codes {
		codes[i].QRImageURL = fmt.Sprintf("%s/api/qrcodes/%s/image", h.baseURL, codes[i].ID)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":     codes,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func (h *QRCodeHandler) GenerateQRCodes(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	batchID := chi.URLParam(r, "id")

	batch, err := dbpkg.GetBatchByID(h.db, batchID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if batch == nil {
		writeError(w, http.StatusNotFound, "batch not found")
		return
	}

	userBaseURL, _ := dbpkg.GetUserBaseURL(h.db, userID)

	var req struct {
		Count       int    `json:"count"`
		LabelPrefix string `json:"label_prefix"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Count < 1 || req.Count > 1000 {
		writeError(w, http.StatusBadRequest, "count must be between 1 and 1000")
		return
	}

	scanBase := effectiveBaseURL(batch.BaseURL, userBaseURL, h.baseURL)
	codes := make([]models.QRCode, 0, req.Count)
	cacheEntries := make(map[string]string)

	for i := 0; i < req.Count; i++ {
		var hash string
		for attempt := 0; attempt < 10; attempt++ {
			h2, err := auth.GenerateHash()
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			if _, exists := h.cache.Get(h2); !exists {
				hash = h2
				break
			}
		}
		if hash == "" {
			writeError(w, http.StatusInternalServerError, "failed to generate unique hash")
			return
		}

		code := models.QRCode{
			ID:           uuid.New().String(),
			Hash:         hash,
			BatchID:      batchID,
			EffectiveURL: batch.RedirectURL,
		}
		if req.LabelPrefix != "" {
			l := fmt.Sprintf("%s-%d", req.LabelPrefix, i+1)
			code.Label = &l
		}

		codes = append(codes, code)
		cacheEntries[hash] = batch.RedirectURL
	}

	if err := dbpkg.CreateQRCodes(h.db, codes); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.cache.BulkSet(cacheEntries)

	for i := range codes {
		codes[i].QRImageURL = fmt.Sprintf("%s/api/qrcodes/%s/image", h.baseURL, codes[i].ID)
		codes[i].ScanURL = fmt.Sprintf("%s/r/%s", scanBase, codes[i].Hash)
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"qr_codes": codes,
		"created":  len(codes),
	})
}

func (h *QRCodeHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	code, err := dbpkg.GetQRCodeByIDForUser(h.db, id, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if code == nil {
		writeError(w, http.StatusNotFound, "QR code not found")
		return
	}
	code.QRImageURL = fmt.Sprintf("%s/api/qrcodes/%s/image", h.baseURL, code.ID)
	writeJSON(w, http.StatusOK, code)
}

func (h *QRCodeHandler) UpdateQRCode(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	var rawReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawReq); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var redirectURL **string
	var label *string

	if v, ok := rawReq["redirect_url"]; ok {
		inner := new(*string)
		if v == nil {
			*inner = nil
		} else if s, ok := v.(string); ok {
			*inner = &s
		}
		redirectURL = inner
	}
	if v, ok := rawReq["label"]; ok {
		if s, ok := v.(string); ok {
			label = &s
		}
	}

	code, err := dbpkg.UpdateQRCode(h.db, id, userID, redirectURL, label)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if code == nil {
		writeError(w, http.StatusNotFound, "QR code not found")
		return
	}

	h.cache.Set(code.Hash, code.EffectiveURL)
	code.QRImageURL = fmt.Sprintf("%s/api/qrcodes/%s/image", h.baseURL, code.ID)
	writeJSON(w, http.StatusOK, code)
}

func (h *QRCodeHandler) DeleteQRCode(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	code, err := dbpkg.GetQRCodeByIDForUser(h.db, id, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if code == nil {
		writeError(w, http.StatusNotFound, "QR code not found")
		return
	}

	if err := dbpkg.DeleteQRCode(h.db, id, userID); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "QR code not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.cache.Delete(code.Hash)
	w.WriteHeader(http.StatusNoContent)
}

func (h *QRCodeHandler) GetQRCodeImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size <= 0 || size > 1024 {
		size = 256
	}

	info, err := dbpkg.GetQRCodeImageInfo(h.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if info == nil {
		writeError(w, http.StatusNotFound, "QR code not found")
		return
	}

	scanBase := effectiveBaseURL(info.BatchBaseURL, info.UserBaseURL, h.baseURL)
	png, err := qr.GeneratePNG(info.Hash, scanBase, size)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate QR code")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `inline; filename="qr.png"`)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	w.Write(png)
}

func (h *QRCodeHandler) GetQRCodeAnalytics(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	analytics, err := dbpkg.GetQRCodeAnalytics(h.db, id, userID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if analytics == nil {
		writeError(w, http.StatusNotFound, "QR code not found")
		return
	}
	writeJSON(w, http.StatusOK, analytics)
}

func (h *QRCodeHandler) ExportBatchCSV(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	batchID := chi.URLParam(r, "id")

	batch, err := dbpkg.GetBatchByID(h.db, batchID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if batch == nil {
		writeError(w, http.StatusNotFound, "batch not found")
		return
	}

	userBaseURL, _ := dbpkg.GetUserBaseURL(h.db, userID)
	scanBase := effectiveBaseURL(batch.BaseURL, userBaseURL, h.baseURL)

	codes, err := dbpkg.GetAllQRCodesForBatch(h.db, batchID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Sanitise batch name for use in the filename
	safeName := strings.Map(func(r rune) rune {
		if r == ' ' {
			return '_'
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return -1
	}, batch.Name)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.csv"`, safeName))

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"label", "scan_url", "effective_url", "redirect_url", "scan_count", "last_scanned_at", "created_at"})

	for _, code := range codes {
		label := ""
		if code.Label != nil {
			label = *code.Label
		}
		redirectURL := ""
		if code.RedirectURL != nil {
			redirectURL = *code.RedirectURL
		}
		lastScannedAt := ""
		if code.LastScannedAt != nil {
			lastScannedAt = code.LastScannedAt.Format(time.RFC3339)
		}
		_ = cw.Write([]string{
			label,
			fmt.Sprintf("%s/r/%s", scanBase, code.Hash),
			code.EffectiveURL,
			redirectURL,
			strconv.Itoa(code.ScanCount),
			lastScannedAt,
			code.CreatedAt.Format(time.RFC3339),
		})
	}

	cw.Flush()
}
