package geoip

import (
	"net"
	"strings"

	"github.com/ip2location/ip2location-go"
	"github.com/oidc-mytoken/utils/httpclient"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
)

var geoDB *ip2location.DB
var privateCountryCode string

// Init initializes the geo ip db
func Init() {
	db, err := ip2location.OpenDB(config.Get().GeoIPDBFile)
	if err != nil {
		log.WithError(err).Error()
		return
	}
	log.Debug("Loaded geo ip data")
	if geoDB != nil {
		geoDB.Close()
	}
	geoDB = db

	res, err := httpclient.Do().R().Get("http://ip-api.com/line/?fields=query")
	if err != nil {
		log.WithError(err).Error("error while retrieving public ip")
	}
	myIP := strings.TrimSpace(string(res.Body()))
	log.WithField("public_ip", myIP).Debug("Obtained public ip")
	privateCountryCode = CountryCode(myIP)
	log.WithField("private_country_code", privateCountryCode).Debug("Determined server location")
}

// CountryCode returns the country code string for a given ip
func CountryCode(ip string) string {
	netIP := net.ParseIP(ip)
	if netIP != nil && (netIP.IsPrivate() || netIP.IsLoopback()) {
		return privateCountryCode
	}
	res, _ := geoDB.Get_country_short(ip)
	return res.Country_short
}
