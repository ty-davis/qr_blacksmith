package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/qr-blacksmith/backend/internal/auth"
	dbpkg "github.com/qr-blacksmith/backend/internal/db"
)

type AuthHandler struct {
	db         *sql.DB
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthHandler(database *sql.DB, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthHandler {
	return &AuthHandler{db: database, jwtSecret: jwtSecret, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/auth",
		Expires:  expires,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		writeError(w, http.StatusBadRequest, "invalid email address")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	existing, err := dbpkg.GetUserByEmail(h.db, req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if existing != nil {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	userID := uuid.New().String()
	if err := dbpkg.CreateUser(h.db, userID, req.Email, string(hash)); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	accessToken, err := auth.IssueAccessToken(userID, h.jwtSecret, h.accessTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	raw, hashed, err := auth.IssueRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	expiresAt := time.Now().Add(h.refreshTTL)
	rtID := uuid.New().String()
	if err := dbpkg.CreateRefreshToken(h.db, rtID, userID, hashed, expiresAt); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.setRefreshCookie(w, raw, expiresAt)

	user, _ := dbpkg.GetUserByID(h.db, userID)
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"access_token": accessToken,
		"user":         user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := dbpkg.GetUserByEmail(h.db, req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken, err := auth.IssueAccessToken(user.ID, h.jwtSecret, h.accessTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	raw, hashed, err := auth.IssueRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	expiresAt := time.Now().Add(h.refreshTTL)
	rtID := uuid.New().String()
	if err := dbpkg.CreateRefreshToken(h.db, rtID, user.ID, hashed, expiresAt); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.setRefreshCookie(w, raw, expiresAt)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"user":         user,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	hash := auth.HashRefreshToken(cookie.Value)
	rt, err := dbpkg.GetRefreshTokenByHash(h.db, hash)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if rt == nil || rt.ExpiresAt.Before(time.Now()) {
		h.clearRefreshCookie(w)
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	if err := dbpkg.DeleteRefreshToken(h.db, rt.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	accessToken, err := auth.IssueAccessToken(rt.UserID, h.jwtSecret, h.accessTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	raw, hashed, err := auth.IssueRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	expiresAt := time.Now().Add(h.refreshTTL)
	rtID := uuid.New().String()
	if err := dbpkg.CreateRefreshToken(h.db, rtID, rt.UserID, hashed, expiresAt); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.setRefreshCookie(w, raw, expiresAt)

	user, _ := dbpkg.GetUserByID(h.db, rt.UserID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"user":         user,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		hash := auth.HashRefreshToken(cookie.Value)
		rt, _ := dbpkg.GetRefreshTokenByHash(h.db, hash)
		if rt != nil {
			dbpkg.DeleteRefreshToken(h.db, rt.ID)
		}
	}
	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	user, err := dbpkg.GetUserByID(h.db, userID)
	if err != nil || user == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req struct {
		BaseURL *string `json:"base_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := dbpkg.UpdateUserSettings(h.db, userID, req.BaseURL)
	if err != nil || user == nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, user)
}
