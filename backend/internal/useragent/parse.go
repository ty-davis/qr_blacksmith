package useragent

import "github.com/mileusna/useragent"

type Info struct {
	DeviceType string
	Browser    string
	OS         string
}

func Parse(rawUA string) Info {
	ua := useragent.Parse(rawUA)

	deviceType := "unknown"
	switch {
	case ua.Mobile:
		deviceType = "mobile"
	case ua.Tablet:
		deviceType = "tablet"
	case ua.Desktop:
		deviceType = "desktop"
	case ua.Bot:
		deviceType = "bot"
	}

	return Info{
		DeviceType: deviceType,
		Browser:    ua.Name,
		OS:         ua.OS,
	}
}
