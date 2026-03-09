package geo

import (
	"net"

	"github.com/oschwald/maxminddb-golang"
)

type Resolver struct {
	db *maxminddb.Reader
}

func New(path string) (*Resolver, error) {
	if path == "" {
		return &Resolver{}, nil
	}
	r, err := maxminddb.Open(path)
	if err != nil {
		return &Resolver{}, nil
	}
	return &Resolver{db: r}, nil
}

type mmRecord struct {
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Country struct {
		Names   map[string]string `maxminddb:"names"`
		ISOCode string            `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

func (r *Resolver) Lookup(ip string) (city, country, countryCode string) {
	if r.db == nil {
		return "", "", ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", "", ""
	}
	var rec mmRecord
	if err := r.db.Lookup(parsed, &rec); err != nil {
		return "", "", ""
	}
	return rec.City.Names["en"], rec.Country.Names["en"], rec.Country.ISOCode
}

func (r *Resolver) Close() {
	if r.db != nil {
		r.db.Close()
	}
}
