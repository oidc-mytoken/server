package utils

import (
	"fmt"
	"strings"

	"github.com/fatih/structs"
	"github.com/oidc-mytoken/utils/utils/issuerutils"

	"github.com/oidc-mytoken/server/internal/utils/hashutils"
)

// CreateMytokenSubject creates the subject of a Mytoken from the oidc subject and oidc issuer
func CreateMytokenSubject(oidcSub, oidcIss string) string {
	comb := issuerutils.CombineSubIss(oidcSub, oidcIss)
	return hashutils.SHA3_256Str([]byte(comb))
}

// CompareNullableIntsWithNilAsInfinity compare two *int64 and handles nil as infinity. It returns 0 if both are equal,
// a positive value if a is greater than b, a negative value if a is less than b
func CompareNullableIntsWithNilAsInfinity(a, b *int64) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil { // b!=nil
		return 1
	}
	if b == nil { // a!=nil
		return -1
	}
	// a and b != nil
	if *a == *b {
		return 0
	} else if *a > *b {
		return 1
	} else {
		return -1
	}
}

// SplitIgnoreEmpty splits a string at the specified delimiter without generating empty parts
func SplitIgnoreEmpty(s, del string) (ret []string) {
	tmp := strings.Split(s, del)
	for _, ss := range tmp {
		if ss != "" {
			ret = append(ret, ss)
		}
	}
	return
}

// RSplitN splits a string s at the delimiter del into n pieces. Unlike strings.SplitN RSplitN splits the string
// starting from the right side
func RSplitN(s, del string, n int) []string {
	if n == 0 {
		return nil
	}
	if del == "" {
		return nil
	}
	if n < 0 {
		return strings.Split(s, del)
	}
	split := make([]string, n)
	delLen := len(del)
	n--
	for n > 0 {
		m := strings.LastIndex(s, del)
		if m < 0 {
			break
		}
		split[n] = s[m+delLen:]
		s = s[:m+delLen-1]
		n--
	}
	split[n] = s
	return split[n:]
}

// StructToStringMap creates a string map from an interface{} using the passed tag name
func StructToStringMap(st interface{}, tag string) map[string]string {
	s := structs.New(st)
	s.TagName = tag
	m := make(map[string]string)
	for k, v := range s.Map() {
		var str string
		switch v := v.(type) {
		case string:
			str = v
		default:
			str = fmt.Sprintf("%v", v)
		}
		m[k] = str
	}
	return m
}

// StructToStringMapUsingJSONTags creates a string map from an interface{} using json tags
func StructToStringMapUsingJSONTags(st interface{}) map[string]string {
	return StructToStringMap(st, "json")
}

// ORErrors returns the first passed error that is not nil
func ORErrors(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// OR logically ORs multiple bools
func OR(bools ...bool) bool {
	for _, b := range bools {
		if b {
			return b
		}
	}
	return false
}

// GetStringFromAnyMap returns a string value from a map[string]any
func GetStringFromAnyMap(m map[string]any, key string) string {
	v, found := m[key]
	if !found {
		return ""
	}
	s, _ := v.(string)
	return s
}

// GetBoolFromAnyMap returns a bool value from a map[string]any
func GetBoolFromAnyMap(m map[string]any, key string) bool {
	v, found := m[key]
	if !found {
		return false
	}
	b, _ := v.(bool)
	return b
}
