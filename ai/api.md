# API Contract

Base URL: `https://yourdomain.com`

All `/api/` endpoints return `Content-Type: application/json`.
Errors return `{ "error": "description" }` with an appropriate HTTP status code.

**Authentication:** All `/api/batches`, `/api/qrcodes`, and `/api/analytics` endpoints
require an `Authorization: Bearer <access_token>` header. Omitting or sending an
expired token returns `401 Unauthorized`. The `/r/:hash` redirect and `/api/auth/*`
endpoints are public.

---

## Auth Endpoints (Public)

### `POST /api/auth/register`
Create a new user account.

**Request:**
```json
{ "email": "user@example.com", "password": "minimum8chars" }
```

**Response:** `201 Created`
```json
{
  "access_token": "<jwt>",
  "user": { "id": "uuid", "email": "user@example.com", "created_at": "..." }
}
```
Sets an `HttpOnly` refresh token cookie.

---

### `POST /api/auth/login`
Authenticate with email + password.

**Request:**
```json
{ "email": "user@example.com", "password": "..." }
```

**Response:** `200 OK`
```json
{
  "access_token": "<jwt>",
  "user": { "id": "uuid", "email": "user@example.com", "created_at": "..." }
}
```
Sets an `HttpOnly` refresh token cookie. Returns `401` on bad credentials
(same message for unknown email and wrong password to prevent enumeration).

---

### `POST /api/auth/refresh`
Exchange a valid refresh token cookie for a new access token.

No request body. Reads the `HttpOnly` cookie automatically.

**Response:** `200 OK`
```json
{ "access_token": "<new_jwt>" }
```
Also rotates the refresh token (old cookie invalidated, new one set).
Returns `401` if the cookie is missing, expired, or revoked.

---

### `POST /api/auth/logout`
Revoke the current refresh token.

No request body. `Authorization` header optional (best-effort).

**Response:** `204 No Content`. Clears the cookie.

---

### `GET /api/me`
Get the authenticated user's profile. Requires `Authorization` header.

**Response:** `200 OK`
```json
{ "id": "uuid", "email": "user@example.com", "created_at": "..." }
```

---

## Redirect Endpoint (Public)

### `GET /r/:hash`

The hot path. Looks up the hash, logs the scan asynchronously, and redirects.

| Status | Meaning |
|---|---|
| `302 Found` | Redirect to effective URL |
| `404 Not Found` | Hash does not exist |

No request body. No auth required. Optimized for minimum latency.

---

## Batches

### `GET /api/batches`
List all batches, most recently created first.

**Query params:**
- `page` (default: 1)
- `per_page` (default: 20, max: 100)

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Summer Campaign",
      "description": "Posters for the summer launch",
      "redirect_url": "https://example.com/summer",
      "qr_code_count": 50,
      "total_scans": 1240,
      "created_at": "2026-03-01T12:00:00Z",
      "updated_at": "2026-03-01T12:00:00Z"
    }
  ],
  "total": 5,
  "page": 1,
  "per_page": 20
}
```

---

### `POST /api/batches`
Create a new batch.

**Request:**
```json
{
  "name": "Summer Campaign",
  "description": "Optional description",
  "redirect_url": "https://example.com/summer"
}
```

**Response:** `201 Created` with the created batch object.

---

### `GET /api/batches/:id`
Get a single batch with summary stats.

**Response:** Single batch object (same shape as list item above).

---

### `PUT /api/batches/:id`
Update batch name, description, or redirect URL.

**Request:** (all fields optional)
```json
{
  "name": "New Name",
  "description": "Updated description",
  "redirect_url": "https://example.com/new-destination"
}
```

**Response:** `200 OK` with updated batch object.

---

### `PUT /api/batches/:id/redirect`
Update the redirect URL for a batch with control over individual overrides.

**Request:**
```json
{
  "redirect_url": "https://example.com/new-destination",
  "override_individuals": false
}
```

- `override_individuals: false` — only codes with no personal override are affected
- `override_individuals: true` — all codes in the batch are updated, individual
  overrides are cleared

**Response:** `200 OK`
```json
{
  "updated_batch": true,
  "qr_codes_updated": 48,
  "qr_codes_skipped": 2
}
```

---

### `DELETE /api/batches/:id`
Delete a batch and all its QR codes and scan data (CASCADE).

**Response:** `204 No Content`

---

### `GET /api/batches/:id/analytics`
Aggregate analytics for all QR codes in a batch.

**Query params:**
- `from` — ISO 8601 datetime (default: 30 days ago)
- `to`   — ISO 8601 datetime (default: now)

**Response:**
```json
{
  "total_scans": 1240,
  "unique_codes_scanned": 38,
  "scans_by_day": [
    { "date": "2026-03-01", "count": 42 }
  ],
  "top_countries": [
    { "country": "United States", "country_code": "US", "count": 800 }
  ],
  "device_breakdown": [
    { "device_type": "mobile", "count": 950 },
    { "device_type": "desktop", "count": 290 }
  ]
}
```

---

## QR Codes

### `GET /api/batches/:id/qrcodes`
List all QR codes in a batch.

**Query params:**
- `page`, `per_page`

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "hash": "aB3xYz9qR2mN4w",
      "batch_id": "uuid",
      "redirect_url": null,
      "effective_url": "https://example.com/summer",
      "label": "Flyer A",
      "scan_count": 24,
      "last_scanned_at": "2026-03-08T14:22:00Z",
      "qr_image_url": "/api/qrcodes/uuid/image",
      "created_at": "2026-03-01T12:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "per_page": 20
}
```

---

### `POST /api/batches/:id/qrcodes`
Generate N QR codes for a batch.

**Request:**
```json
{
  "count": 25,
  "label_prefix": "Flyer"
}
```

`label_prefix` is optional. If provided, codes are labelled "Flyer 1", "Flyer 2", etc.

**Response:** `201 Created`
```json
{
  "created": 25,
  "qr_codes": [ /* array of QR code objects */ ]
}
```

---

### `GET /api/qrcodes/:id`
Get a single QR code with stats.

---

### `PUT /api/qrcodes/:id`
Update a single QR code's redirect URL or label.

**Request:**
```json
{
  "redirect_url": "https://example.com/specific-page",
  "label": "Bus Stop Poster"
}
```

Set `redirect_url` to `null` to clear the override and revert to batch URL.

**Response:** `200 OK` with updated QR code object.

---

### `DELETE /api/qrcodes/:id`
Delete a single QR code and its scan history.

**Response:** `204 No Content`

---

### `GET /api/qrcodes/:id/image`
Returns the QR code as a PNG image.

**Query params:**
- `size` — pixel size of the image (default: 256, max: 1024)

**Response:** `Content-Type: image/png`

---

### `GET /api/qrcodes/:id/analytics`
Analytics for a single QR code.

**Query params:** `from`, `to` (same as batch analytics)

**Response:** Same shape as batch analytics but scoped to one code.

---

## Dashboard

### `GET /api/analytics/overview`
Top-level stats for the dashboard home page.

**Response:**
```json
{
  "total_batches": 5,
  "total_qr_codes": 120,
  "total_scans": 4821,
  "scans_today": 38,
  "scans_this_week": 210,
  "most_scanned_code": { /* QR code object */ },
  "recent_scans": [ /* last 10 scan events with code info */ ]
}
```
