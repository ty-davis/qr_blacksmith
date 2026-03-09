# QR Blacksmith — Project Overview

## What It Does
A QR code tracking platform. Users create batches of QR codes, deploy them in the
physical world (print, stickers, packaging, etc.), and then view analytics on when
and where each code was scanned.

When someone scans a code, they are silently redirected to the configured destination
URL. The backend logs the scan (timestamp, city, device info) asynchronously so the
redirect itself is as fast as possible.

---

## Repository Layout

```
qr_ninja/
├── ai/               ← Planning documents (this folder)
├── backend/          ← Go API server
└── frontend/         ← Vue 3 + TypeScript + Tailwind CSS
```

---

## Core Concepts

### User
An account identified by email + password. Every resource (batches, QR codes, scan data)
belongs to a user. Authentication uses short-lived JWT access tokens (15 min) plus
long-lived refresh tokens (30 days) stored in an `HttpOnly` cookie.

### Batch
A named group of QR codes. Every batch has a `redirect_url` that all codes in it
will use by default. You can change the batch redirect URL at any time.

### QR Code
An individual scannable code. Each has a unique, cryptographically random hash baked
into a URL like `https://yourdomain.com/r/<hash>`. A QR code can have its own
`redirect_url` that overrides the batch-level one.

### Redirect Resolution
```
effective_url = qr_code.redirect_url ?? batch.redirect_url
```

### Bulk Update (Option 3)
When updating a batch's redirect URL, the caller chooses:
- `override_individuals: false` — only codes with no personal override are updated
- `override_individuals: true`  — all codes in the batch get the new URL, clearing
                                   any individual overrides

### Scan
An event recorded each time a QR code is scanned. Contains:
- Timestamp
- City and country (derived from IP via GeoIP lookup — raw IP is never stored)
- Device type (mobile/desktop/tablet)
- Browser name
- OS name
- Raw user agent string (retained for future re-parsing)

---

## Quality & Security Goals

| Goal | Approach |
|---|---|
| Fast redirects | In-memory hash→URL cache in Go + async scan logging |
| Unguessable codes | 10 bytes from `crypto/rand` → 14-char base64url (2^80 entropy) |
| No PII stored | GeoIP lookup then IP discarded; no name/email collected |
| Concurrent safety | SQLite in WAL mode; RWMutex on in-memory cache |
| Resilience | Buffered channel for scan events; graceful drain on shutdown |
| Open redirect protection | Destination URLs validated/owned by the account |
| Authentication | JWT access tokens (15 min) + refresh tokens (30 days, HttpOnly cookie) |
| Password storage | bcrypt (cost 12) — plaintext password never persisted |
| Resource isolation | All batch/QR queries filter by `user_id` from JWT claim — users never see each other's data |
| Public scan path | `/r/:hash` requires no auth — scanners are anonymous members of the public |

---

## See Also
- [database.md](./database.md) — full schema with rationale
- [backend.md](./backend.md) — Go server architecture and API endpoints
- [frontend.md](./frontend.md) — Vue app structure and views
- [api.md](./api.md) — full REST API contract
