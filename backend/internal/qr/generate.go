package qr

import (
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

func GeneratePNG(hash, baseURL string, size int) ([]byte, error) {
	if size <= 0 || size > 1024 {
		size = 256
	}
	url := fmt.Sprintf("%s/r/%s", baseURL, hash)
	return qrcode.Encode(url, qrcode.Medium, size)
}
