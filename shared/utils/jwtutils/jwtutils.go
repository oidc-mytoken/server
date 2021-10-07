package jwtutils

import (
	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
)

// ResultSet is a slice of values with an additional value indicating if it was set
type ResultSet []struct {
	Value interface{}
	Set   bool
}

// GetFromJWT returns the values for the requested keys from the JWT
func GetFromJWT(token string, key ...string) (values ResultSet) {
	if atJWT, _ := jwt.Parse(token, nil); atJWT != nil {
		log.Trace("Parsed token")
		if claims, ok := atJWT.Claims.(jwt.MapClaims); ok {
			for _, k := range key {
				v, set := claims[k]
				values = append(
					values, struct {
						Value interface{}
						Set   bool
					}{
						Value: v,
						Set:   set,
					},
				)
			}
		}
	}
	return values
}

// GetStringFromJWT returns a string value for the given key
func GetStringFromJWT(token, key string) (string, bool) {
	res := GetFromJWT(token, key)
	if len(res) != 1 {
		return "", false
	}
	if !res[0].Set {
		return "", false
	}
	v, ok := res[0].Value.(string)
	return v, ok
}

// GetAudiencesFromJWT parses the passed jwt token and returns the aud claim as a slice of strings
func GetAudiencesFromJWT(token string) ([]string, bool) {
	log.Trace("Getting auds from token")
	res := GetFromJWT(token, "aud")
	if len(res) != 1 {
		return nil, false
	}
	if !res[0].Set {
		return nil, false
	}
	auds := res[0].Value
	switch v := auds.(type) {
	case string:
		return []string{v}, true
	case []string:
		return v, true
	case []interface{}:
		strs := []string{}
		for _, s := range v {
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
