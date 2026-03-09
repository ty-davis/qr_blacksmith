CREATE TABLE scans (
    id           TEXT PRIMARY KEY,
    qr_code_id   TEXT NOT NULL REFERENCES qr_codes(id) ON DELETE CASCADE,
    scanned_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    city         TEXT,
    country      TEXT,
    country_code TEXT,
    device_type  TEXT,
    browser      TEXT,
    os           TEXT,
    user_agent   TEXT NOT NULL
);
CREATE INDEX idx_scans_qr_code_id ON scans(qr_code_id);
CREATE INDEX idx_scans_scanned_at ON scans(scanned_at);
