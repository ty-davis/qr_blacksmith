package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/qr-blacksmith/backend/internal/auth"
	"github.com/qr-blacksmith/backend/internal/cache"
	"github.com/qr-blacksmith/backend/internal/db"
	"github.com/qr-blacksmith/backend/internal/geo"
	"github.com/qr-blacksmith/backend/internal/handlers"
)

func main() {
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./data/qr_blacksmith.db")
	geoIPPath := getEnv("GEOIP_DB_PATH", "")
	baseURL := getEnv("BASE_URL", "http://localhost:8080")
	corsOrigin := getEnv("CORS_ORIGIN", "http://localhost:5173")
	jwtSecret := getEnv("JWT_SECRET", "change-me-please-use-a-real-secret")

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		accessTTL = 15 * time.Minute
	}
	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "720h"))
	if err != nil {
		refreshTTL = 720 * time.Hour
	}

	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("failed to create data dir: %v", err)
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer database.Close()

	geoRes, err := geo.New(geoIPPath)
	if err != nil {
		log.Printf("geo resolver init error: %v", err)
		geoRes, _ = geo.New("")
	}
	defer geoRes.Close()

	redirectCache := cache.New()
	entries, err := db.LoadAllRedirectEntries(database)
	if err != nil {
		log.Printf("failed to load redirect cache: %v", err)
	} else {
		redirectCache.BulkSet(entries)
		log.Printf("loaded %d redirect entries into cache", len(entries))
	}

	scanCh := make(chan handlers.ScanEvent, 2000)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go handlers.ScanWorker(database, geoRes, scanCh, &wg)
	}

	authHandler := handlers.NewAuthHandler(database, jwtSecret, accessTTL, refreshTTL)
	batchHandler := handlers.NewBatchHandler(database, redirectCache)
	qrCodeHandler := handlers.NewQRCodeHandler(database, redirectCache, baseURL)
	redirectHandler := handlers.NewRedirectHandler(database, redirectCache, scanCh)
	scanHandler := handlers.NewScanHandler(database)

	r := chi.NewRouter()
	r.Use(handlers.Recovery)
	r.Use(handlers.Logger)
	r.Use(handlers.CORS(corsOrigin))

	// Public routes
	r.Get("/r/{hash}", redirectHandler.ServeHTTP)
	r.Get("/api/qrcodes/{id}/image", qrCodeHandler.GetQRCodeImage)
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/refresh", authHandler.Refresh)
	r.Post("/api/auth/logout", authHandler.Logout)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth(jwtSecret))

		r.Get("/api/auth/me", authHandler.Me)
		r.Patch("/api/auth/me", authHandler.UpdateMe)

		r.Get("/api/batches", batchHandler.ListBatches)
		r.Post("/api/batches", batchHandler.CreateBatch)
		r.Get("/api/batches/{id}", batchHandler.GetBatch)
		r.Put("/api/batches/{id}", batchHandler.UpdateBatch)
		r.Delete("/api/batches/{id}", batchHandler.DeleteBatch)
		r.Put("/api/batches/{id}/redirect", batchHandler.UpdateBatchRedirect)
		r.Get("/api/batches/{id}/analytics", batchHandler.GetBatchAnalytics)
		r.Get("/api/batches/{id}/qrcodes", qrCodeHandler.ListQRCodes)
		r.Post("/api/batches/{id}/qrcodes", qrCodeHandler.GenerateQRCodes)

		r.Get("/api/qrcodes/{id}", qrCodeHandler.GetQRCode)
		r.Put("/api/qrcodes/{id}", qrCodeHandler.UpdateQRCode)
		r.Delete("/api/qrcodes/{id}", qrCodeHandler.DeleteQRCode)
		r.Get("/api/qrcodes/{id}/analytics", qrCodeHandler.GetQRCodeAnalytics)

		r.Get("/api/analytics/overview", scanHandler.GetOverviewStats)
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}

		close(scanCh)
		wg.Wait()
		log.Println("shutdown complete")
	}()

	log.Printf("QR Blacksmith server starting on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
