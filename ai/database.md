# Database Plan — SQLite

## Setup Pragmas (applied at every connection open)

```sql
PRAGMA journal_mode = WAL;       -- concurrent reads during writes
PRAGMA foreign_keys = ON;        -- enforce FK constraints
PRAGMA busy_timeout = 5000;      -- wait up to 5s on lock instead of erroring
PRAGMA synchronous = NORMAL;     -- safe with WAL, faster than FULL
```

WAL mode is critical here because the redirect handler (reads) and the scan logger
(writes) run concurrently. Without WAL, writes would block all reads.

---

## Schema

### `users`
```sql
CREATE TABLE users (
    id              TEXT PRIMARY KEY,            -- UUID v4
    email           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,               -- bcrypt cost 12
    created_at      DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at      DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_users_email ON users(email);
```

### `refresh_tokens`
```sql
CREATE TABLE refresh_tokens (
    id          TEXT PRIMARY KEY,               -- UUID v4 (also used as the token itself)
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  DATETIME NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
```

**Notes:**
- Refresh tokens are stored hashed (SHA-256) so a DB leak cannot be used to hijack sessions
- On logout, the token row is deleted (single device) or all rows for `user_id` are deleted (all devices)
- Expired tokens are cleaned up by a background goroutine (e.g. every hour)

### `batches`
```sql
CREATE TABLE batches (
    id           TEXT PRIMARY KEY,            -- UUID v4
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    description  TEXT,
    redirect_url TEXT NOT NULL,               -- default destination for all codes
    created_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_batches_user ON batches(user_id);
```

### `qr_codes`
```sql
CREATE TABLE qr_codes (
    id           TEXT PRIMARY KEY,            -- UUID v4
    hash         TEXT NOT NULL UNIQUE,        -- 14-char base64url, crypto/rand
    batch_id     TEXT NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    redirect_url TEXT,                        -- NULL = inherit from batch
    label        TEXT,                        -- optional human label e.g. "Flyer A"
    created_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_qr_codes_hash    ON qr_codes(hash);
CREATE INDEX idx_qr_codes_batch   ON qr_codes(batch_id);
```

**Notes:**
- `hash` is indexed as it is hit on every single redirect request
- `label` lets users annotate individual codes (e.g. "Bus stop poster", "Instagram bio")
- `redirect_url` being NULL means "use batch URL" — explicit NULL, not empty string

### `scans`
```sql
CREATE TABLE scans (
    id           TEXT PRIMARY KEY,            -- UUID v4
    qr_code_id   TEXT NOT NULL REFERENCES qr_codes(id) ON DELETE CASCADE,
    scanned_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    city         TEXT,                        -- may be NULL if GeoIP lookup fails
    country      TEXT,                        -- full name e.g. "United States"
    country_code TEXT,                        -- ISO 3166-1 alpha-2 e.g. "US"
    device_type  TEXT,                        -- "mobile" | "desktop" | "tablet" | "unknown"
    browser      TEXT,                        -- e.g. "Chrome", "Safari"
    os           TEXT,                        -- e.g. "iOS", "Android", "Windows"
    user_agent   TEXT NOT NULL               -- raw string, retained for re-parsing
);

CREATE INDEX idx_scans_qr_code_id ON scans(qr_code_id);
CREATE INDEX idx_scans_scanned_at ON scans(scanned_at);
```

**Privacy notes:**
- Raw IP address is looked up via GeoIP library then immediately discarded
- Only city + country are persisted — no IP, no precise coordinates
- `user_agent` is retained because parsing libraries improve over time and raw UA
  lets you re-derive device info without re-collecting data

---

## Migrations

Use `golang-migrate/migrate` with numbered SQL files:

```
backend/migrations/
  001_create_users.up.sql
  001_create_users.down.sql
  002_create_refresh_tokens.up.sql
  002_create_refresh_tokens.down.sql
  003_create_batches.up.sql
  003_create_batches.down.sql
  004_create_qr_codes.up.sql
  004_create_qr_codes.down.sql
  005_create_scans.up.sql
  005_create_scans.down.sql
```

Run migrations automatically on server startup before accepting connections.

---

## Indexes — Rationale

| Index | Why |
|---|---|
| `users(email)` | Login lookup — must be fast |
| `refresh_tokens(user_id)` | Logout (delete all tokens for a user) |
| `batches(user_id)` | Listing all batches for the authenticated user |
| `qr_codes(hash)` | Hit on every redirect — must be sub-millisecond |
| `qr_codes(batch_id)` | Listing all codes in a batch, bulk updates |
| `scans(qr_code_id)` | Per-code analytics queries |
| `scans(scanned_at)` | Time-series charts, filtering by date range |

---

## Analytics Query Patterns

These are the queries the frontend will need to support:

```sql
-- Total scans for a QR code
SELECT COUNT(*) FROM scans WHERE qr_code_id = ?;

-- Scans per day for the last 30 days
SELECT DATE(scanned_at) as day, COUNT(*) as count
FROM scans
WHERE qr_code_id = ? AND scanned_at >= datetime('now', '-30 days')
GROUP BY day ORDER BY day;

-- Top countries for a batch
SELECT country, country_code, COUNT(*) as count
FROM scans
JOIN qr_codes ON scans.qr_code_id = qr_codes.id
WHERE qr_codes.batch_id = ?
GROUP BY country ORDER BY count DESC LIMIT 10;

-- Device type breakdown
SELECT device_type, COUNT(*) as count
FROM scans WHERE qr_code_id = ?
GROUP BY device_type;

-- Batch-level scan total
SELECT COUNT(*) FROM scans
JOIN qr_codes ON scans.qr_code_id = qr_codes.id
WHERE qr_codes.batch_id = ?;
```

---

## Soft Delete Consideration

Currently using hard deletes with CASCADE. If you want audit history (e.g. "this
code was deleted but had 500 scans"), add `deleted_at DATETIME` to `batches` and
`qr_codes` and filter all queries with `WHERE deleted_at IS NULL`.

Recommendation: Start with hard deletes (simpler), add soft deletes later if needed.
