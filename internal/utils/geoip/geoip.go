package geoip

import (
	"github.com/ip2location/ip2location-go"
	log "github.com/sirupsen/logrus"
)

var geoDB *ip2location.DB

func init() {
	db, err := ip2location.OpenDB("./IP2LOCATION-LITE-DB1.IPV6.BIN")
	if err != nil {
		log.WithError(err).Error()
	}
	log.Debug("Loaded geo ip data")
	geoDB = db
}

// Country returns the country name string for a given ip
func Country(ip string) string {
	res, _ := geoDB.Get_country_long(ip)
	return res.Country_long
}

// CountryCode returns the country code string for a given ip
func CountryCode(ip string) string {
	res, _ := geoDB.Get_country_short(ip)
	return res.Country_short
}
