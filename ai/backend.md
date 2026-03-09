# Backend Plan — Go

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              -- entry point: config, wire up, start
├── internal/
│   ├── auth/
│   │   ├── jwt.go               -- issue/validate access tokens (JWT HS256)
│   │   └── middleware.go        -- RequireAuth: extract + validate Bearer token, inject userID into ctx
│   ├── cache/
│   │   └── redirect.go          -- in-memory hash→URL cache with RWMutex
│   ├── db/
│   │   ├── db.go                -- open connection, apply pragmas, run migrations
│   │   └── queries.go           -- all raw SQL query functions
│   ├── geo/
│   │   └── geo.go               -- MaxMind GeoLite2 IP→city lookup
│   ├── handlers/
│   │   ├── auth.go              -- register, login, refresh, logout, me
│   │   ├── batches.go           -- CRUD for batches (user-scoped)
│   │   ├── qrcodes.go           -- CRUD for QR codes, image generation
│   │   ├── redirect.go          -- the hot path: hash lookup → redirect
│   │   ├── scans.go             -- scan analytics endpoints
│   │   └── middleware.go        -- CORS, logging, recovery
│   ├── models/
│   │   └── models.go            -- shared Go structs (User, Batch, QRCode, Scan)
│   ├── qr/
│   │   └── generate.go          -- QR code PNG/SVG generation
│   └── useragent/
│       └── parse.go             -- parse raw UA → device/browser/os
├── migrations/
│   ├── 001_create_users.up.sql
│   ├── 001_create_users.down.sql
│   ├── 002_create_refresh_tokens.up.sql
│   ├── 002_create_refresh_tokens.down.sql
│   ├── 003_create_batches.up.sql
│   ├── 003_create_batches.down.sql
│   ├── 004_create_qr_codes.up.sql
│   ├── 004_create_qr_codes.down.sql
│   ├── 005_create_scans.up.sql
│   └── 005_create_scans.down.sql
├── data/
│   └── GeoLite2-City.mmdb        -- MaxMind DB (gitignored)
├── .env.example
├── go.mod
└── go.sum
```

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/go-chi/chi/v5` | HTTP router (lightweight, idiomatic) |
| `modernc.org/sqlite` | Pure-Go SQLite driver (no CGO required) |
| `github.com/golang-migrate/migrate/v4` | SQL migration runner |
| `github.com/google/uuid` | UUID v4 generation for IDs |
| `github.com/skip2/go-qrcode` | QR code PNG image generation |
| `github.com/oschwald/maxminddb-golang` | MaxMind GeoLite2 database reader |
| `github.com/mileusna/useragent` | User agent string parsing |
| `github.com/joho/godotenv` | .env file loading |
| `golang.org/x/crypto/bcrypt` | Password hashing (cost 12) |
| `github.com/golang-jwt/jwt/v5` | JWT access token signing and validation |

---

## Configuration (`.env`)

```env
PORT=8080
DB_PATH=./data/qr_blacksmith.db
GEOIP_DB_PATH=./data/GeoLite2-City.mmdb
BASE_URL=https://yourdomain.com       # used to build QR code URLs
CORS_ORIGIN=http://localhost:5173     # Vue dev server
JWT_SECRET=change-me-to-a-random-64-byte-secret
JWT_ACCESS_TTL=15m                    # access token lifetime
JWT_REFRESH_TTL=720h                  # refresh token lifetime (30 days)
```

---

## Authentication

### Token Strategy

| Token | TTL | Transport |
|---|---|---|
| Access token | 15 min | `Authorization: Bearer <token>` header |
| Refresh token | 30 days | `HttpOnly; Secure; SameSite=Strict` cookie |

Access tokens are JWT (HS256) signed with `JWT_SECRET`. The payload carries `user_id` and `exp`.
Refresh tokens are opaque UUIDs stored **hashed** (SHA-256) in the `refresh_tokens` table.

### JWT Package (`internal/auth/jwt.go`)

```go
type Claims struct {
    UserID string `json:"user_id"`
    jwt.RegisteredClaims
}

func IssueAccessToken(userID string, secret string, ttl time.Duration) (string, error)
func ValidateAccessToken(tokenStr string, secret string) (*Claims, error)
func IssueRefreshToken() (raw string, hashed string, err error)
    // raw → sent to client as cookie
    // hashed → stored in DB
```

### Auth Middleware (`internal/auth/middleware.go`)

```go
// RequireAuth validates the Bearer token and injects userID into request context.
// Returns 401 if missing or invalid, 401 if expired.
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler

// UserIDFromCtx retrieves the user ID injected by RequireAuth.
func UserIDFromCtx(ctx context.Context) string
```

All `/api/batches`, `/api/qrcodes`, and `/api/analytics` routes are wrapped with `RequireAuth`.
`/r/:hash` and `/api/auth/*` are public.

### Resource Ownership

Every query for batches or QR codes includes a `user_id` filter derived from the JWT claim:

```go
userID := auth.UserIDFromCtx(r.Context())
batch, err := db.GetBatchByIDForUser(ctx, batchID, userID)
// returns 404 if batch exists but belongs to a different user
```

This prevents horizontal privilege escalation — users never see each other's data.

---

## The Redirect Handler (Critical Path)

This is the most important endpoint. Every QR scan hits it.

### Flow
```
GET /r/:hash
  1. Look up hash in RedirectCache (RWMutex read lock — microseconds)
  2. If miss → query SQLite, populate cache, continue
  3. If hash not found → 404
  4. Send scan event to buffered channel (non-blocking)
  5. Return HTTP 302 to redirect URL
```

### In-Memory Cache

```go
type RedirectCache struct {
    mu    sync.RWMutex
    items map[string]string  // hash → effective redirect URL
}

func (c *RedirectCache) Get(hash string) (string, bool)
func (c *RedirectCache) Set(hash string, url string)
func (c *RedirectCache) Delete(hash string)
func (c *RedirectCache) BulkSet(entries map[string]string)
```

Populated at startup by loading all active codes. Updated whenever a redirect URL
is changed. Never expires (codes don't change frequently).

### Async Scan Logger

```go
type ScanEvent struct {
    QRCodeID  string
    IP        string      // used for GeoIP lookup, then discarded
    UserAgent string
    Time      time.Time
}

// Buffered channel — redirect handler never blocks on scan write
scanCh := make(chan ScanEvent, 2000)

// Worker pool (e.g. 4 goroutines) reads from channel, writes to SQLite
func scanWorker(db *sql.DB, geo *geo.Resolver, ch <-chan ScanEvent)
```

On graceful shutdown: close channel, wait for workers to drain before exit.

---

## Hash Generation

```go
import (
    "crypto/rand"
    "encoding/base64"
)

func generateHash() (string, error) {
    b := make([]byte, 10)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(b), nil
    // produces a 14-character URL-safe string
    // entropy: 2^80 combinations — effectively unguessable
}
```

Retry loop on the extremely unlikely collision: generate → check DB → if exists, retry.

---

## Bulk Redirect Update Logic

```go
type UpdateBatchRedirectRequest struct {
    RedirectURL         string `json:"redirect_url"`
    OverrideIndividuals bool   `json:"override_individuals"`
}
```

**If `override_individuals = false`:**
```sql
UPDATE batches SET redirect_url = ?, updated_at = ? WHERE id = ?;
-- QR codes with their own redirect_url are untouched
-- Cache: only update entries for codes with redirect_url IS NULL
```

**If `override_individuals = true`:**
```sql
UPDATE batches SET redirect_url = ?, updated_at = ? WHERE id = ?;
UPDATE qr_codes SET redirect_url = NULL, updated_at = ? WHERE batch_id = ?;
-- Cache: update ALL entries for this batch to new URL
```

---

## QR Code Image Generation

Each QR code embeds a URL: `{BASE_URL}/r/{hash}`

Images are generated **on demand** (not stored on disk):
- `GET /api/qrcodes/:id/image?format=png&size=300` — returns PNG bytes
- `GET /api/qrcodes/:id/image?format=svg` — returns SVG (future)

No caching needed at this scale; generation is fast. Add file caching later if load increases.

---

## GeoIP Lookup

Using MaxMind GeoLite2-City (free, requires account registration):

```go
func (r *Resolver) Lookup(ip string) (city, country, countryCode string) {
    // Parse IP
    // Open record from mmdb
    // Extract city name (English), country name, ISO code
    // Return empty strings on any error — scan is still recorded
}
```

The raw IP is passed to this function and immediately goes out of scope. It is never
written to any log file or database.

Note: GeoLite2-City.mmdb must be downloaded separately and is gitignored. Document
the download step in README.

---

## Middleware Stack

```
Recovery (panic → 500)
  └── Request Logger (method, path, status, duration — no IPs logged)
        └── CORS (configured origin from env)
              └── Public routes (/r/:hash, /api/auth/*)
              └── RequireAuth (JWT validation)
                    └── Protected routes (/api/batches, /api/qrcodes, /api/analytics/*)
```

---

## Graceful Shutdown

```go
// On SIGINT / SIGTERM:
// 1. Stop accepting new connections (http.Server.Shutdown)
// 2. Close scan channel
// 3. Wait for scan workers to drain (WaitGroup)
// 4. Close DB connection
// 5. Exit 0
```

This ensures no scan events are lost when the process is killed.

---

## Rate Limiting (Future)

Not in v1, but add to the redirect endpoint before going to production:
- Per-IP: max 60 requests/minute to `/r/` (prevents scan flooding / fake analytics)
- Use a token bucket in memory or Redis if multi-instance

---

## Error Handling Principles

- Never log or return raw IP addresses
- All handler errors return JSON: `{"error": "message"}`
- 404 on unknown hash (don't hint at whether hash format is wrong)
- 500 errors should not leak internal details to client
