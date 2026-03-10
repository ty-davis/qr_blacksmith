package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/qr-blacksmith/backend/internal/models"
)

func parseTime(s string) time.Time {
	formats := []string{
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

// ─── Users ────────────────────────────────────────────────────────────────────

func CreateUser(db *sql.DB, id, email, passwordHash string) error {
	_, err := db.Exec(
		`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		id, email, passwordHash,
	)
	return err
}

func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	row := db.QueryRow(
		`SELECT id, email, password_hash, base_url, created_at, updated_at FROM users WHERE email = ?`, email,
	)
	return scanUser(row)
}

func GetUserByID(db *sql.DB, id string) (*models.User, error) {
	row := db.QueryRow(
		`SELECT id, email, password_hash, base_url, created_at, updated_at FROM users WHERE id = ?`, id,
	)
	return scanUser(row)
}

func scanUser(row *sql.Row) (*models.User, error) {
	u := &models.User{}
	var baseURL sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &baseURL, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if baseURL.Valid && baseURL.String != "" {
		u.BaseURL = &baseURL.String
	}
	u.CreatedAt = parseTime(createdAt)
	u.UpdatedAt = parseTime(updatedAt)
	return u, nil
}

// UpdateUserSettings updates mutable account-level fields.
// Pass nil to leave a field unchanged; pass a pointer to "" to clear it.
func UpdateUserSettings(db *sql.DB, userID string, baseURL *string) (*models.User, error) {
	var val interface{}
	if baseURL == nil {
		// no-op for base_url — read current and return
		return GetUserByID(db, userID)
	}
	if *baseURL == "" {
		val = nil // clear
	} else {
		val = *baseURL
	}
	_, err := db.Exec(
		`UPDATE users SET base_url = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = ?`,
		val, userID,
	)
	if err != nil {
		return nil, err
	}
	return GetUserByID(db, userID)
}

// GetUserBaseURL returns only the user's custom base_url (nil if not set).
func GetUserBaseURL(db *sql.DB, userID string) (*string, error) {
	var v sql.NullString
	err := db.QueryRow(`SELECT base_url FROM users WHERE id = ?`, userID).Scan(&v)
	if err != nil {
		return nil, err
	}
	if !v.Valid || v.String == "" {
		return nil, nil
	}
	return &v.String, nil
}

// ─── Refresh Tokens ───────────────────────────────────────────────────────────

func CreateRefreshToken(db *sql.DB, id, userID, tokenHash string, expiresAt time.Time) error {
	_, err := db.Exec(
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at) VALUES (?, ?, ?, ?)`,
		id, userID, tokenHash, expiresAt.UTC().Format(time.RFC3339),
	)
	return err
}

func GetRefreshTokenByHash(db *sql.DB, hash string) (*models.RefreshToken, error) {
	row := db.QueryRow(
		`SELECT id, user_id, token_hash, expires_at, created_at FROM refresh_tokens WHERE token_hash = ?`,
		hash,
	)
	rt := &models.RefreshToken{}
	var expiresAt, createdAt string
	err := row.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &expiresAt, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rt.ExpiresAt = parseTime(expiresAt)
	rt.CreatedAt = parseTime(createdAt)
	return rt, nil
}

func DeleteRefreshToken(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM refresh_tokens WHERE id = ?`, id)
	return err
}

func DeleteAllRefreshTokensForUser(db *sql.DB, userID string) error {
	_, err := db.Exec(`DELETE FROM refresh_tokens WHERE user_id = ?`, userID)
	return err
}

func DeleteExpiredRefreshTokens(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < ?`, time.Now().UTC().Format(time.RFC3339))
	return err
}

// ─── Batches ──────────────────────────────────────────────────────────────────

func CreateBatch(db *sql.DB, batch *models.Batch) error {
	var baseURL interface{}
	if batch.BaseURL != nil && *batch.BaseURL != "" {
		baseURL = *batch.BaseURL
	}
	_, err := db.Exec(
		`INSERT INTO batches (id, user_id, name, description, redirect_url, base_url) VALUES (?, ?, ?, ?, ?, ?)`,
		batch.ID, batch.UserID, batch.Name, batch.Description, batch.RedirectURL, baseURL,
	)
	return err
}

func GetBatchByID(db *sql.DB, id, userID string) (*models.Batch, error) {
	row := db.QueryRow(`
		SELECT b.id, b.user_id, b.name, b.description, b.redirect_url, b.base_url,
			COUNT(DISTINCT qc.id) as qr_code_count,
			COUNT(s.id) as total_scans,
			b.created_at, b.updated_at
		FROM batches b
		LEFT JOIN qr_codes qc ON qc.batch_id = b.id
		LEFT JOIN scans s ON s.qr_code_id = qc.id
		WHERE b.id = ? AND b.user_id = ?
		GROUP BY b.id
	`, id, userID)
	return scanBatch(row)
}

func scanBatch(row *sql.Row) (*models.Batch, error) {
	b := &models.Batch{}
	var desc, baseURL sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(
		&b.ID, &b.UserID, &b.Name, &desc, &b.RedirectURL, &baseURL,
		&b.QRCodeCount, &b.TotalScans,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if desc.Valid {
		b.Description = &desc.String
	}
	if baseURL.Valid && baseURL.String != "" {
		b.BaseURL = &baseURL.String
	}
	b.CreatedAt = parseTime(createdAt)
	b.UpdatedAt = parseTime(updatedAt)
	return b, nil
}

func ListBatches(db *sql.DB, userID string, page, perPage int) ([]models.Batch, int, error) {
	offset := (page - 1) * perPage

	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM batches WHERE user_id = ?`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.Query(`
		SELECT b.id, b.user_id, b.name, b.description, b.redirect_url, b.base_url,
			COUNT(DISTINCT qc.id) as qr_code_count,
			COUNT(s.id) as total_scans,
			b.created_at, b.updated_at
		FROM batches b
		LEFT JOIN qr_codes qc ON qc.batch_id = b.id
		LEFT JOIN scans s ON s.qr_code_id = qc.id
		WHERE b.user_id = ?
		GROUP BY b.id
		ORDER BY b.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var batches []models.Batch
	for rows.Next() {
		var b models.Batch
		var desc, baseURL sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.Name, &desc, &b.RedirectURL, &baseURL,
			&b.QRCodeCount, &b.TotalScans,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, 0, err
		}
		if desc.Valid {
			b.Description = &desc.String
		}
		if baseURL.Valid && baseURL.String != "" {
			b.BaseURL = &baseURL.String
		}
		b.CreatedAt = parseTime(createdAt)
		b.UpdatedAt = parseTime(updatedAt)
		batches = append(batches, b)
	}

	return batches, total, rows.Err()
}

// UpdateBatch updates mutable batch fields. Pass nil to leave a field unchanged;
// for base_url pass a pointer to "" to clear it.
func UpdateBatch(db *sql.DB, id, userID string, name, description *string, redirectURL *string, baseURL *string) (*models.Batch, error) {
	sets := []string{"updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')"}
	args := []interface{}{}

	if name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *name)
	}
	if description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *description)
	}
	if redirectURL != nil {
		sets = append(sets, "redirect_url = ?")
		args = append(args, *redirectURL)
	}
	if baseURL != nil {
		sets = append(sets, "base_url = ?")
		if *baseURL == "" {
			args = append(args, nil) // clear
		} else {
			args = append(args, *baseURL)
		}
	}

	args = append(args, id, userID)
	_, err := db.Exec(
		fmt.Sprintf("UPDATE batches SET %s WHERE id = ? AND user_id = ?", strings.Join(sets, ", ")),
		args...,
	)
	if err != nil {
		return nil, err
	}
	return GetBatchByID(db, id, userID)
}

func DeleteBatch(db *sql.DB, id, userID string) error {
	result, err := db.Exec(`DELETE FROM batches WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func UpdateBatchRedirect(db *sql.DB, id, userID, redirectURL string, overrideIndividuals bool) (codesUpdated, codesSkipped int, err error) {
	tx, txErr := db.Begin()
	if txErr != nil {
		return 0, 0, txErr
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE batches SET redirect_url = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = ? AND user_id = ?`,
		redirectURL, id, userID,
	)
	if err != nil {
		return
	}

	var total int
	if err = tx.QueryRow(`SELECT COUNT(*) FROM qr_codes WHERE batch_id = ?`, id).Scan(&total); err != nil {
		return
	}

	if overrideIndividuals {
		result, e := tx.Exec(`UPDATE qr_codes SET redirect_url = NULL WHERE batch_id = ?`, id)
		if e != nil {
			err = e
			return
		}
		n, _ := result.RowsAffected()
		codesUpdated = int(n)
	} else {
		if err = tx.QueryRow(`SELECT COUNT(*) FROM qr_codes WHERE batch_id = ? AND redirect_url IS NOT NULL`, id).Scan(&codesSkipped); err != nil {
			return
		}
		codesUpdated = total - codesSkipped
	}

	err = tx.Commit()
	return
}

// ─── QR Codes ─────────────────────────────────────────────────────────────────

func CreateQRCodes(db *sql.DB, codes []models.QRCode) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO qr_codes (id, hash, batch_id, redirect_url, label) VALUES (?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, code := range codes {
		if _, err := stmt.Exec(code.ID, code.Hash, code.BatchID, code.RedirectURL, code.Label); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func GetQRCodeByIDForUser(db *sql.DB, id, userID string) (*models.QRCode, error) {
	row := db.QueryRow(`
		SELECT qc.id, qc.hash, qc.batch_id, qc.redirect_url, qc.label,
			COALESCE(qc.redirect_url, b.redirect_url) as effective_url,
			COUNT(s.id) as scan_count,
			MAX(s.scanned_at) as last_scanned_at,
			qc.created_at
		FROM qr_codes qc
		JOIN batches b ON qc.batch_id = b.id
		LEFT JOIN scans s ON s.qr_code_id = qc.id
		WHERE qc.id = ? AND b.user_id = ?
		GROUP BY qc.id
	`, id, userID)
	return scanQRCode(row)
}

// GetQRCodeByIDPublic looks up a QR code by ID without user ownership check.
// Used for public endpoints (e.g. serving QR image PNGs).
func GetQRCodeByIDPublic(db *sql.DB, id string) (*models.QRCode, error) {
	row := db.QueryRow(`
		SELECT qc.id, qc.hash, qc.batch_id, qc.redirect_url, qc.label,
			COALESCE(qc.redirect_url, b.redirect_url) as effective_url,
			0 as scan_count,
			NULL as last_scanned_at,
			qc.created_at
		FROM qr_codes qc
		JOIN batches b ON qc.batch_id = b.id
		WHERE qc.id = ?
	`, id)
	return scanQRCode(row)
}

// QRCodeImageInfo carries the minimal data needed to render a QR code image.
type QRCodeImageInfo struct {
	Hash         string
	BatchBaseURL *string
	UserBaseURL  *string
}

// GetQRCodeImageInfo returns the hash and applicable base URLs for a QR code
// without any ownership check (used for the public image endpoint).
func GetQRCodeImageInfo(db *sql.DB, id string) (*QRCodeImageInfo, error) {
	var hash string
	var bBU, uBU sql.NullString
	err := db.QueryRow(`
		SELECT qc.hash, b.base_url, u.base_url
		FROM qr_codes qc
		JOIN batches b ON qc.batch_id = b.id
		JOIN users u ON b.user_id = u.id
		WHERE qc.id = ?
	`, id).Scan(&hash, &bBU, &uBU)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	info := &QRCodeImageInfo{Hash: hash}
	if bBU.Valid && bBU.String != "" {
		info.BatchBaseURL = &bBU.String
	}
	if uBU.Valid && uBU.String != "" {
		info.UserBaseURL = &uBU.String
	}
	return info, nil
}

func GetQRCodeByHash(db *sql.DB, hash string) (*models.QRCode, error) {
	row := db.QueryRow(`
		SELECT qc.id, qc.hash, qc.batch_id, qc.redirect_url, qc.label,
			COALESCE(qc.redirect_url, b.redirect_url) as effective_url,
			0 as scan_count,
			NULL as last_scanned_at,
			qc.created_at
		FROM qr_codes qc
		JOIN batches b ON qc.batch_id = b.id
		WHERE qc.hash = ?
	`, hash)
	return scanQRCode(row)
}

func scanQRCode(row *sql.Row) (*models.QRCode, error) {
	qc := &models.QRCode{}
	var redirectURL, label sql.NullString
	var lastScannedAt sql.NullString
	var createdAt string
	err := row.Scan(
		&qc.ID, &qc.Hash, &qc.BatchID, &redirectURL, &label,
		&qc.EffectiveURL,
		&qc.ScanCount, &lastScannedAt,
		&createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if redirectURL.Valid {
		qc.RedirectURL = &redirectURL.String
	}
	if label.Valid {
		qc.Label = &label.String
	}
	if lastScannedAt.Valid && lastScannedAt.String != "" {
		t := parseTime(lastScannedAt.String)
		qc.LastScannedAt = &t
	}
	qc.CreatedAt = parseTime(createdAt)
	return qc, nil
}

func ListQRCodes(db *sql.DB, batchID, userID string, page, perPage int) ([]models.QRCode, int, error) {
	offset := (page - 1) * perPage

	var exists int
	if err := db.QueryRow(`SELECT COUNT(*) FROM batches WHERE id = ? AND user_id = ?`, batchID, userID).Scan(&exists); err != nil {
		return nil, 0, err
	}
	if exists == 0 {
		return nil, 0, nil
	}

	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM qr_codes WHERE batch_id = ?`, batchID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.Query(`
		SELECT qc.id, qc.hash, qc.batch_id, qc.redirect_url, qc.label,
			COALESCE(qc.redirect_url, b.redirect_url) as effective_url,
			COUNT(s.id) as scan_count,
			MAX(s.scanned_at) as last_scanned_at,
			qc.created_at
		FROM qr_codes qc
		JOIN batches b ON qc.batch_id = b.id
		LEFT JOIN scans s ON s.qr_code_id = qc.id
		WHERE qc.batch_id = ?
		GROUP BY qc.id
		ORDER BY qc.created_at DESC
		LIMIT ? OFFSET ?
	`, batchID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var codes []models.QRCode
	for rows.Next() {
		var qc models.QRCode
		var redirectURL, label sql.NullString
		var lastScannedAt sql.NullString
		var createdAt string
		if err := rows.Scan(
			&qc.ID, &qc.Hash, &qc.BatchID, &redirectURL, &label,
			&qc.EffectiveURL,
			&qc.ScanCount, &lastScannedAt,
			&createdAt,
		); err != nil {
			return nil, 0, err
		}
		if redirectURL.Valid {
			qc.RedirectURL = &redirectURL.String
		}
		if label.Valid {
			qc.Label = &label.String
		}
		if lastScannedAt.Valid && lastScannedAt.String != "" {
			t := parseTime(lastScannedAt.String)
			qc.LastScannedAt = &t
		}
		qc.CreatedAt = parseTime(createdAt)
		codes = append(codes, qc)
	}

	return codes, total, rows.Err()
}

func UpdateQRCode(db *sql.DB, id, userID string, redirectURL **string, label *string) (*models.QRCode, error) {
	sets := []string{"updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')"}
	args := []interface{}{}

	if redirectURL != nil {
		sets = append(sets, "redirect_url = ?")
		args = append(args, *redirectURL) // *redirectURL can be nil (sets to NULL)
	}
	if label != nil {
		sets = append(sets, "label = ?")
		args = append(args, *label)
	}

	args = append(args, id, userID)
	_, err := db.Exec(
		fmt.Sprintf(
			`UPDATE qr_codes SET %s WHERE id = ? AND batch_id IN (SELECT id FROM batches WHERE user_id = ?)`,
			strings.Join(sets, ", "),
		),
		args...,
	)
	if err != nil {
		return nil, err
	}
	return GetQRCodeByIDForUser(db, id, userID)
}

func DeleteQRCode(db *sql.DB, id, userID string) error {
	result, err := db.Exec(
		`DELETE FROM qr_codes WHERE id = ? AND batch_id IN (SELECT id FROM batches WHERE user_id = ?)`,
		id, userID,
	)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func LoadAllRedirectEntries(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query(`
		SELECT qc.hash, COALESCE(qc.redirect_url, b.redirect_url)
		FROM qr_codes qc
		JOIN batches b ON qc.batch_id = b.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make(map[string]string)
	for rows.Next() {
		var hash, url string
		if err := rows.Scan(&hash, &url); err != nil {
			return nil, err
		}
		entries[hash] = url
	}
	return entries, rows.Err()
}

// ─── Scans ────────────────────────────────────────────────────────────────────

func CreateScan(db *sql.DB, scan *models.Scan) error {
	_, err := db.Exec(
		`INSERT INTO scans (id, qr_code_id, city, country, country_code, device_type, browser, os, user_agent)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		scan.ID, scan.QRCodeID, scan.City, scan.Country, scan.CountryCode,
		scan.DeviceType, scan.Browser, scan.OS, scan.UserAgent,
	)
	return err
}

func buildTimeFilter(condition, from, to string, args []interface{}) (string, []interface{}) {
	if from != "" {
		condition += " AND s.scanned_at >= ?"
		args = append(args, from)
	}
	if to != "" {
		condition += " AND s.scanned_at <= ?"
		args = append(args, to)
	}
	return condition, args
}

func fillAnalytics(db *sql.DB, baseCondition string, args []interface{}) (*models.ScanAnalytics, error) {
	analytics := &models.ScanAnalytics{}

	if err := db.QueryRow(
		fmt.Sprintf(`SELECT COUNT(*) FROM scans s WHERE %s`, baseCondition), args...,
	).Scan(&analytics.TotalScans); err != nil {
		return nil, err
	}

	rows, err := db.Query(
		fmt.Sprintf(`SELECT date(s.scanned_at) as day, COUNT(*) FROM scans s WHERE %s GROUP BY day ORDER BY day`, baseCondition),
		args...,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var d models.DayScanCount
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			rows.Close()
			return nil, err
		}
		analytics.ScansByDay = append(analytics.ScansByDay, d)
	}
	rows.Close()

	rows2, err := db.Query(
		fmt.Sprintf(`SELECT COALESCE(s.country, 'Unknown'), COALESCE(s.country_code, ''), COUNT(*) as cnt FROM scans s WHERE %s GROUP BY s.country_code ORDER BY cnt DESC LIMIT 10`, baseCondition),
		args...,
	)
	if err != nil {
		return nil, err
	}
	for rows2.Next() {
		var c models.CountryScanCount
		if err := rows2.Scan(&c.Country, &c.CountryCode, &c.Count); err != nil {
			rows2.Close()
			return nil, err
		}
		analytics.TopCountries = append(analytics.TopCountries, c)
	}
	rows2.Close()

	rows3, err := db.Query(
		fmt.Sprintf(`SELECT COALESCE(s.device_type, 'unknown'), COUNT(*) as cnt FROM scans s WHERE %s GROUP BY s.device_type ORDER BY cnt DESC`, baseCondition),
		args...,
	)
	if err != nil {
		return nil, err
	}
	for rows3.Next() {
		var d models.DeviceScanCount
		if err := rows3.Scan(&d.DeviceType, &d.Count); err != nil {
			rows3.Close()
			return nil, err
		}
		analytics.DeviceBreakdown = append(analytics.DeviceBreakdown, d)
	}
	rows3.Close()

	if analytics.ScansByDay == nil {
		analytics.ScansByDay = []models.DayScanCount{}
	}
	if analytics.TopCountries == nil {
		analytics.TopCountries = []models.CountryScanCount{}
	}
	if analytics.DeviceBreakdown == nil {
		analytics.DeviceBreakdown = []models.DeviceScanCount{}
	}

	return analytics, nil
}

func GetBatchAnalytics(db *sql.DB, batchID, userID, from, to string) (*models.ScanAnalytics, error) {
	var exists int
	if err := db.QueryRow(`SELECT COUNT(*) FROM batches WHERE id = ? AND user_id = ?`, batchID, userID).Scan(&exists); err != nil || exists == 0 {
		return nil, nil
	}

	condition := `s.qr_code_id IN (SELECT id FROM qr_codes WHERE batch_id = ?)`
	args := []interface{}{batchID}
	condition, args = buildTimeFilter(condition, from, to, args)
	return fillAnalytics(db, condition, args)
}

func GetQRCodeAnalytics(db *sql.DB, qrCodeID, userID, from, to string) (*models.ScanAnalytics, error) {
	var exists int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM qr_codes qc JOIN batches b ON qc.batch_id = b.id WHERE qc.id = ? AND b.user_id = ?`,
		qrCodeID, userID,
	).Scan(&exists); err != nil || exists == 0 {
		return nil, nil
	}

	condition := `s.qr_code_id = ?`
	args := []interface{}{qrCodeID}
	condition, args = buildTimeFilter(condition, from, to, args)
	return fillAnalytics(db, condition, args)
}

func GetOverviewStats(db *sql.DB, userID string) (*models.OverviewStats, error) {
	stats := &models.OverviewStats{}

	if err := db.QueryRow(`SELECT COUNT(*) FROM batches WHERE user_id = ?`, userID).Scan(&stats.TotalBatches); err != nil {
		return nil, err
	}
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM qr_codes qc JOIN batches b ON qc.batch_id = b.id WHERE b.user_id = ?`, userID,
	).Scan(&stats.TotalQRCodes); err != nil {
		return nil, err
	}
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM scans s
		JOIN qr_codes qc ON s.qr_code_id = qc.id
		JOIN batches b ON qc.batch_id = b.id
		WHERE b.user_id = ?`, userID,
	).Scan(&stats.TotalScans); err != nil {
		return nil, err
	}
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM scans s
		JOIN qr_codes qc ON s.qr_code_id = qc.id
		JOIN batches b ON qc.batch_id = b.id
		WHERE b.user_id = ? AND date(s.scanned_at) = date('now')`, userID,
	).Scan(&stats.ScansToday); err != nil {
		return nil, err
	}
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM scans s
		JOIN qr_codes qc ON s.qr_code_id = qc.id
		JOIN batches b ON qc.batch_id = b.id
		WHERE b.user_id = ? AND s.scanned_at >= datetime('now', '-7 days')`, userID,
	).Scan(&stats.ScansThisWeek); err != nil {
		return nil, err
	}

	// Most scanned code
	row := db.QueryRow(`
		SELECT qc.id, qc.hash, qc.batch_id, qc.redirect_url, qc.label,
			COALESCE(qc.redirect_url, b.redirect_url) as effective_url,
			COUNT(s.id) as scan_count,
			MAX(s.scanned_at) as last_scanned_at,
			qc.created_at
		FROM qr_codes qc
		JOIN batches b ON qc.batch_id = b.id
		LEFT JOIN scans s ON s.qr_code_id = qc.id
		WHERE b.user_id = ?
		GROUP BY qc.id
		ORDER BY scan_count DESC
		LIMIT 1
	`, userID)
	mostScanned, err := scanQRCode(row)
	if err != nil {
		return nil, err
	}
	stats.MostScannedCode = mostScanned

	// Recent scans
	rows, err := db.Query(`
		SELECT s.qr_code_id, qc.label, b.name, s.scanned_at, s.city, s.country, COALESCE(s.device_type, 'unknown')
		FROM scans s
		JOIN qr_codes qc ON s.qr_code_id = qc.id
		JOIN batches b ON qc.batch_id = b.id
		WHERE b.user_id = ?
		ORDER BY s.scanned_at DESC
		LIMIT 10
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rs models.RecentScan
		var label, city, country sql.NullString
		var scannedAt string
		if err := rows.Scan(&rs.QRCodeID, &label, &rs.BatchName, &scannedAt, &city, &country, &rs.DeviceType); err != nil {
			return nil, err
		}
		rs.ScannedAt = parseTime(scannedAt)
		if label.Valid {
			rs.QRCodeLabel = &label.String
		}
		if city.Valid {
			rs.City = &city.String
		}
		if country.Valid {
			rs.Country = &country.String
		}
		stats.RecentScans = append(stats.RecentScans, rs)
	}

	if stats.RecentScans == nil {
		stats.RecentScans = []models.RecentScan{}
	}

	return stats, rows.Err()
}

// GetAllQRCodesForBatch returns every QR code in a batch without pagination,
// used for CSV export. Returns nil if the batch doesn't belong to userID.
func GetAllQRCodesForBatch(db *sql.DB, batchID, userID string) ([]models.QRCode, error) {
	var exists int
	if err := db.QueryRow(`SELECT COUNT(*) FROM batches WHERE id = ? AND user_id = ?`, batchID, userID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, nil
	}

	rows, err := db.Query(`
		SELECT qc.id, qc.hash, qc.batch_id, qc.redirect_url, qc.label,
			COALESCE(qc.redirect_url, b.redirect_url) as effective_url,
			COUNT(s.id) as scan_count,
			MAX(s.scanned_at) as last_scanned_at,
			qc.created_at
		FROM qr_codes qc
		JOIN batches b ON qc.batch_id = b.id
		LEFT JOIN scans s ON s.qr_code_id = qc.id
		WHERE qc.batch_id = ?
		GROUP BY qc.id
		ORDER BY qc.created_at ASC
	`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []models.QRCode
	for rows.Next() {
		var qc models.QRCode
		var redirectURL, label sql.NullString
		var lastScannedAt sql.NullString
		var createdAt string
		if err := rows.Scan(
			&qc.ID, &qc.Hash, &qc.BatchID, &redirectURL, &label,
			&qc.EffectiveURL,
			&qc.ScanCount, &lastScannedAt,
			&createdAt,
		); err != nil {
			return nil, err
		}
		if redirectURL.Valid {
			qc.RedirectURL = &redirectURL.String
		}
		if label.Valid {
			qc.Label = &label.String
		}
		if lastScannedAt.Valid && lastScannedAt.String != "" {
			t := parseTime(lastScannedAt.String)
			qc.LastScannedAt = &t
		}
		qc.CreatedAt = parseTime(createdAt)
		codes = append(codes, qc)
	}

	return codes, rows.Err()
}
