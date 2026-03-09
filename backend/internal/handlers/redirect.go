package handlers

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/qr-blacksmith/backend/internal/cache"
	dbpkg "github.com/qr-blacksmith/backend/internal/db"
	"github.com/qr-blacksmith/backend/internal/geo"
	"github.com/qr-blacksmith/backend/internal/models"
	"github.com/qr-blacksmith/backend/internal/useragent"
)

// ScanEvent is sent to the async scan worker channel on each QR scan.
type ScanEvent struct {
	Hash      string
	IP        string
	UserAgent string
	Time      time.Time
}

type RedirectHandler struct {
	db     *sql.DB
	cache  *cache.RedirectCache
	scanCh chan<- ScanEvent
}

func NewRedirectHandler(database *sql.DB, c *cache.RedirectCache, scanCh chan<- ScanEvent) *RedirectHandler {
	return &RedirectHandler{db: database, cache: c, scanCh: scanCh}
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	targetURL, ok := h.cache.Get(hash)
	if !ok {
		qrCode, err := dbpkg.GetQRCodeByHash(h.db, hash)
		if err != nil || qrCode == nil {
			http.NotFound(w, r)
			return
		}
		targetURL = qrCode.EffectiveURL
		h.cache.Set(hash, targetURL)
	}

	select {
	case h.scanCh <- ScanEvent{
		Hash:      hash,
		IP:        realIP(r),
		UserAgent: r.UserAgent(),
		Time:      time.Now(),
	}:
	default:
	}

	http.Redirect(w, r, targetURL, http.StatusFound)
}

// ScanWorker processes scan events from the channel asynchronously.
func ScanWorker(database *sql.DB, geoRes *geo.Resolver, ch <-chan ScanEvent, wg *sync.WaitGroup) {
	defer wg.Done()
	for event := range ch {
		qrCode, err := dbpkg.GetQRCodeByHash(database, event.Hash)
		if err != nil || qrCode == nil {
			continue
		}

		city, country, cc := geoRes.Lookup(event.IP)
		ua := useragent.Parse(event.UserAgent)

		scan := &models.Scan{
			ID:         uuid.New().String(),
			QRCodeID:   qrCode.ID,
			UserAgent:  event.UserAgent,
			DeviceType: ua.DeviceType,
			Browser:    ua.Browser,
			OS:         ua.OS,
		}
		if city != "" {
			scan.City = &city
		}
		if country != "" {
			scan.Country = &country
		}
		if cc != "" {
			scan.CountryCode = &cc
		}

		if err := dbpkg.CreateScan(database, scan); err != nil {
			log.Printf("scan write error: %v", err)
		}
	}
}

func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
