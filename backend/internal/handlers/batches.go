package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/qr-blacksmith/backend/internal/auth"
	"github.com/qr-blacksmith/backend/internal/cache"
	dbpkg "github.com/qr-blacksmith/backend/internal/db"
	"github.com/qr-blacksmith/backend/internal/models"
)

type BatchHandler struct {
	db    *sql.DB
	cache *cache.RedirectCache
}

func NewBatchHandler(database *sql.DB, c *cache.RedirectCache) *BatchHandler {
	return &BatchHandler{db: database, cache: c}
}

func (h *BatchHandler) ListBatches(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	batches, total, err := dbpkg.ListBatches(h.db, userID, page, perPage)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if batches == nil {
		batches = []models.Batch{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":     batches,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func (h *BatchHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		RedirectURL string  `json:"redirect_url"`
		BaseURL     *string `json:"base_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.RedirectURL == "" {
		writeError(w, http.StatusBadRequest, "redirect_url is required")
		return
	}
	if _, err := url.ParseRequestURI(req.RedirectURL); err != nil {
		writeError(w, http.StatusBadRequest, "invalid redirect_url")
		return
	}
	if req.BaseURL != nil && *req.BaseURL != "" {
		if _, err := url.ParseRequestURI(*req.BaseURL); err != nil {
			writeError(w, http.StatusBadRequest, "invalid base_url")
			return
		}
	}

	batch := &models.Batch{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		RedirectURL: req.RedirectURL,
		BaseURL:     req.BaseURL,
	}
	if err := dbpkg.CreateBatch(h.db, batch); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	created, _ := dbpkg.GetBatchByID(h.db, batch.ID, userID)
	writeJSON(w, http.StatusCreated, created)
}

func (h *BatchHandler) GetBatch(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	batch, err := dbpkg.GetBatchByID(h.db, id, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if batch == nil {
		writeError(w, http.StatusNotFound, "batch not found")
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func (h *BatchHandler) UpdateBatch(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		RedirectURL *string `json:"redirect_url"`
		BaseURL     *string `json:"base_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RedirectURL != nil {
		if _, err := url.ParseRequestURI(*req.RedirectURL); err != nil {
			writeError(w, http.StatusBadRequest, "invalid redirect_url")
			return
		}
	}
	if req.BaseURL != nil && *req.BaseURL != "" {
		if _, err := url.ParseRequestURI(*req.BaseURL); err != nil {
			writeError(w, http.StatusBadRequest, "invalid base_url")
			return
		}
	}

	batch, err := dbpkg.UpdateBatch(h.db, id, userID, req.Name, req.Description, req.RedirectURL, req.BaseURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if batch == nil {
		writeError(w, http.StatusNotFound, "batch not found")
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func (h *BatchHandler) DeleteBatch(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	if err := dbpkg.DeleteBatch(h.db, id, userID); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "batch not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BatchHandler) UpdateBatchRedirect(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	var req struct {
		RedirectURL         string `json:"redirect_url"`
		OverrideIndividuals bool   `json:"override_individuals"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RedirectURL == "" {
		writeError(w, http.StatusBadRequest, "redirect_url is required")
		return
	}

	updated, skipped, err := dbpkg.UpdateBatchRedirect(h.db, id, userID, req.RedirectURL, req.OverrideIndividuals)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Refresh cache
	entries, _ := dbpkg.LoadAllRedirectEntries(h.db)
	if entries != nil {
		h.cache.BulkSet(entries)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"codes_updated": updated,
		"codes_skipped": skipped,
	})
}

func (h *BatchHandler) GetBatchAnalytics(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	analytics, err := dbpkg.GetBatchAnalytics(h.db, id, userID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if analytics == nil {
		writeError(w, http.StatusNotFound, "batch not found")
		return
	}
	writeJSON(w, http.StatusOK, analytics)
}
