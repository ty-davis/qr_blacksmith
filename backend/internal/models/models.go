package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	BaseURL      *string   `json:"base_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Batch struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	RedirectURL string    `json:"redirect_url"`
	BaseURL     *string   `json:"base_url"`
	QRCodeCount int       `json:"qr_code_count"`
	TotalScans  int       `json:"total_scans"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type QRCode struct {
	ID            string     `json:"id"`
	Hash          string     `json:"hash"`
	BatchID       string     `json:"batch_id"`
	RedirectURL   *string    `json:"redirect_url"`
	EffectiveURL  string     `json:"effective_url"`
	Label         *string    `json:"label"`
	ScanCount     int        `json:"scan_count"`
	LastScannedAt *time.Time `json:"last_scanned_at"`
	QRImageURL    string     `json:"qr_image_url"`
	ScanURL       string     `json:"scan_url"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Scan struct {
	ID          string    `json:"id"`
	QRCodeID    string    `json:"qr_code_id"`
	ScannedAt   time.Time `json:"scanned_at"`
	City        *string   `json:"city"`
	Country     *string   `json:"country"`
	CountryCode *string   `json:"country_code"`
	DeviceType  string    `json:"device_type"`
	Browser     string    `json:"browser"`
	OS          string    `json:"os"`
	UserAgent   string    `json:"user_agent"`
}

type RecentScan struct {
	QRCodeID    string     `json:"qr_code_id"`
	QRCodeLabel *string    `json:"qr_code_label"`
	BatchName   string     `json:"batch_name"`
	ScannedAt   time.Time  `json:"scanned_at"`
	City        *string    `json:"city"`
	Country     *string    `json:"country"`
	DeviceType  string     `json:"device_type"`
}

type ScanAnalytics struct {
	TotalScans         int                `json:"total_scans"`
	UniqueCodesScanned *int               `json:"unique_codes_scanned,omitempty"`
	ScansByDay         []DayScanCount     `json:"scans_by_day"`
	TopCountries       []CountryScanCount `json:"top_countries"`
	DeviceBreakdown    []DeviceScanCount  `json:"device_breakdown"`
}

type DayScanCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type CountryScanCount struct {
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Count       int    `json:"count"`
}

type DeviceScanCount struct {
	DeviceType string `json:"device_type"`
	Count      int    `json:"count"`
}

type OverviewStats struct {
	TotalBatches    int          `json:"total_batches"`
	TotalQRCodes    int          `json:"total_qr_codes"`
	TotalScans      int          `json:"total_scans"`
	ScansToday      int          `json:"scans_today"`
	ScansThisWeek   int          `json:"scans_this_week"`
	MostScannedCode *QRCode      `json:"most_scanned_code"`
	RecentScans     []RecentScan `json:"recent_scans"`
}
