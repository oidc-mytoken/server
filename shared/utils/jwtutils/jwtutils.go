package jwtutils

import (
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

// GetAudiencesFromJWT parses the passed jwt token and returns the aud claim as a slice of strings
func GetAudiencesFromJWT(token string) ([]string, bool) {
	log.Trace("Getting auds from token")
	if atJWT, _ := jwt.Parse(token, nil); atJWT != nil {
		log.Trace("Parsed token")
		if claims, ok := atJWT.Claims.(jwt.MapClaims); ok {
			auds := claims["aud"]
			switch auds.(type) {
			case string:
				return []string{auds.(string)}, true
			case []string:
				return auds.([]string), true
			case []interface{}:
				strs := []string{}
				for _, s := range auds.([]interface{}) {
					str, ok := s.(string)
					if !ok {
						return nil, false
					}
					strs = append(strs, str)
				}
				return strs, true
			default:
				return nil, false
			}
		}
	}
	return nil, false
}
