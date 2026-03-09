CREATE TABLE qr_codes (
    id           TEXT PRIMARY KEY,
    hash         TEXT NOT NULL UNIQUE,
    batch_id     TEXT NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    redirect_url TEXT,
    label        TEXT,
    created_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX idx_qr_codes_hash  ON qr_codes(hash);
CREATE INDEX idx_qr_codes_batch ON qr_codes(batch_id);
