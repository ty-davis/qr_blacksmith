# QR Blacksmith

A self-hosted QR code tracking platform. Generate batch QR codes, track scans with geo and device analytics, and manage redirect URLs — all scoped per user.

## Features

- **Batch management** — organise QR codes into named batches, each with a default redirect URL
- **Bulk generation** — generate up to 1,000 QR codes per batch with optional label prefixes
- **Per-code redirect overrides** — override the batch default on any individual code
- **Custom base URL priority** — scan links use: *batch base URL* → *account base URL* → *server `BASE_URL`*
- **Fast redirect path** — in-memory hash→URL cache for near-zero-latency redirects
- **Scan analytics** — scans by day, top countries, device/browser/OS breakdown
- **Geo enrichment** — optional MaxMind GeoLite2 City database for city/country lookup
- **JWT authentication** — short-lived access tokens (15 min) + rotating `HttpOnly` refresh tokens (30 day)
- **Zero CGO** — uses `modernc.org/sqlite` (pure Go), no C toolchain required

## Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.21, chi router, SQLite (modernc) |
| Frontend | Vue 3 + TypeScript, Vite, Pinia, Tailwind CSS |
| Auth | JWT HS256 access tokens, opaque refresh tokens |
| QR generation | `skip2/go-qrcode` |
| Geo | MaxMind GeoLite2 (optional, fail-open) |

## Project structure

```
qr_blacksmith_2/
├── backend/
│   ├── cmd/server/main.go          # Entry point & router
│   ├── internal/
│   │   ├── auth/                   # JWT issue/validate, hash generation
│   │   ├── cache/                  # In-memory redirect cache (RWMutex)
│   │   ├── db/
│   │   │   ├── db.go               # SQLite open + migration runner
│   │   │   ├── queries.go          # All SQL queries
│   │   │   └── migrations/         # 001–007 up/down SQL files
│   │   ├── geo/                    # MaxMind GeoIP resolver
│   │   ├── handlers/               # HTTP handlers (auth, batches, qrcodes, redirect, scans)
│   │   ├── models/                 # Go structs
│   │   ├── qr/                     # QR PNG generation
│   │   └── useragent/              # UA parsing
│   ├── .env.example
│   └── go.mod
└── frontend/
    └── src/
        ├── api/                    # Axios API clients
        ├── components/             # Reusable Vue components
        ├── router/                 # Vue Router + auth guard
        ├── stores/                 # Pinia stores (auth, batches, qrcodes)
        ├── types/                  # TypeScript interfaces
        └── views/                  # Page components
```

## Getting started

### Prerequisites

- Go 1.21+
- Node.js 18+
- (Optional) [MaxMind GeoLite2 City](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) `.mmdb` file for geo analytics

### Backend

```bash
cd backend
cp .env.example .env
# Edit .env — at minimum set JWT_SECRET to a random value
mkdir -p data
go run ./cmd/server
```

The server starts on `http://localhost:8080`. SQLite migrations run automatically on startup.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The dev server starts on `http://localhost:5173` and proxies `/api` and `/r` to the backend.

### Production build

```bash
# Backend
cd backend && go build -o qr_blacksmith ./cmd/server

# Frontend
cd frontend && npm run build
# Serve dist/ with any static file host, pointing /api and /r at the backend
```

## Configuration (backend/.env)

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `DB_PATH` | `./data/qr_blacksmith.db` | SQLite database file path |
| `GEOIP_DB_PATH` | _(empty)_ | Path to GeoLite2-City.mmdb (optional) |
| `BASE_URL` | `http://localhost:8080` | Server base URL — fallback for QR scan links |
| `CORS_ORIGIN` | `http://localhost:5173` | Allowed CORS origin |
| `JWT_SECRET` | _(placeholder)_ | **Change this** — HS256 signing key |
| `JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `JWT_REFRESH_TTL` | `720h` | Refresh token lifetime (30 days) |

## API overview

### Auth

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/auth/register` | Register a new account |
| `POST` | `/api/auth/login` | Login, receive access token + refresh cookie |
| `POST` | `/api/auth/refresh` | Rotate refresh token, get new access token |
| `POST` | `/api/auth/logout` | Revoke refresh token |
| `GET` | `/api/auth/me` | Get current user |
| `PATCH` | `/api/auth/me` | Update account settings (e.g. `base_url`) |

### Batches

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/batches` | List batches (paginated) |
| `POST` | `/api/batches` | Create batch |
| `GET` | `/api/batches/:id` | Get batch |
| `PUT` | `/api/batches/:id` | Update batch |
| `DELETE` | `/api/batches/:id` | Delete batch |
| `PUT` | `/api/batches/:id/redirect` | Bulk-update redirect URL |
| `GET` | `/api/batches/:id/analytics` | Scan analytics for a batch |

### QR codes

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/batches/:id/qrcodes` | List QR codes in a batch |
| `POST` | `/api/batches/:id/qrcodes` | Generate QR codes (bulk) |
| `GET` | `/api/qrcodes/:id` | Get a single QR code |
| `PUT` | `/api/qrcodes/:id` | Update redirect URL / label |
| `DELETE` | `/api/qrcodes/:id` | Delete QR code |
| `GET` | `/api/qrcodes/:id/image` | Serve QR code PNG (public, cacheable) |
| `GET` | `/api/qrcodes/:id/analytics` | Scan analytics for a code |

### Redirect & analytics

| Method | Path | Description |
|---|---|---|
| `GET` | `/r/:hash` | Redirect (hot path, served from memory) |
| `GET` | `/api/analytics/overview` | Account-wide overview stats |

All routes except `/r/:hash`, `/api/auth/*`, and `/api/qrcodes/:id/image` require a `Authorization: Bearer <token>` header.

## Base URL priority

The domain encoded inside each QR code PNG is resolved in this order:

1. **Batch-level** `base_url` — set when creating or editing a batch
2. **Account-level** `base_url` — set in Account Settings (`PATCH /api/auth/me`)
3. **Server** `BASE_URL` — the `.env` fallback

This lets you use a custom short domain (e.g. `https://go.mycompany.com`) for specific campaigns while the server default handles everything else.

## Database migrations

Migrations live in `backend/internal/db/migrations/` as numbered `*.up.sql` / `*.down.sql` pairs and are embedded into the binary at compile time. They run automatically on server start via a lightweight hand-rolled runner (no CGO-dependent migration libraries).

| # | Description |
|---|---|
| 001 | `users` table |
| 002 | `refresh_tokens` table |
| 003 | `batches` table |
| 004 | `qr_codes` table |
| 005 | `scans` table |
| 006 | `users.base_url` column |
| 007 | `batches.base_url` column |
