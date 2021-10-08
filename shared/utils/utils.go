package utils

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
	"unsafe"

	"github.com/fatih/structs"
)

var src rand.Source

func init() {
	src = rand.NewSource(time.Now().UnixNano())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandASCIIString returns a random string consisting of ASCII characters of the given
// length.
func RandASCIIString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b)) // unsafe is fine here skipcq: GSC-G103
}

// IntersectSlices returns the common elements of two slices
func IntersectSlices(a, b []string) (res []string) {
	for _, bb := range b {
		if StringInSlice(bb, a) {
			res = append(res, bb)
		}
	}
	return
}

// StringInSlice checks if a string is in a slice of strings
func StringInSlice(key string, slice []string) bool {
	for _, s := range slice {
		if s == key {
			return true
		}
	}
	return false
}

// ReplaceStringInSlice replaces all occurrences of a string in a slice with another string
func ReplaceStringInSlice(s *[]string, o, n string, caseSensitive bool) {
	if !caseSensitive {
		o = strings.ToLower(o)
	}
	for i, ss := range *s {
		if !caseSensitive {
			ss = strings.ToLower(ss)
		}
		if o == ss {
			(*s)[i] = n
		}
	}
}

// IsSubSet checks if all strings of a slice 'a' are contained in the slice 'b'
func IsSubSet(a, b []string) bool {
	for _, aa := range a {
		if !StringInSlice(aa, b) {
			return false
		}
	}
	return true
}

// IPsAreSubSet checks if all ips of ipsA are contained in ipsB, it will also check ip subnets
func IPsAreSubSet(ipsA, ipsB []string) bool {
	for _, ipA := range ipsA {
		if !IPIsIn(ipA, ipsB) {
			return false
		}
	}
	return true
}

// IPIsIn checks if a ip is in a slice of ips, it will also check ip subnets
func IPIsIn(ip string, ips []string) bool {
	if len(ips) == 0 {
		return false
	}
	ipA := net.ParseIP(ip)
	for _, ipp := range ips {
		if strings.Contains(ipp, "/") {
			_, ipNetB, _ := net.ParseCIDR(ipp)
			if ipNetB != nil && ipNetB.Contains(ipA) {
				return true
			}
		} else if ip == ipp {
			return true
		}

	}
	return false
}

// CombineURLPath combines multiple parts of a url
func CombineURLPath(p string, ps ...string) (r string) {
	r = p
	for _, pp := range ps {
		if pp == "" {
			continue
		}
		if r == "" {
			r = pp
			continue
		}
		rAppend := r
		ppAppend := pp
		if strings.HasSuffix(r, "/") {
			rAppend = r[:len(r)-1]
		}
		if strings.HasPrefix(pp, "/") {
			ppAppend = pp[1:]
		}
		r = fmt.Sprintf("%s%c%s", rAppend, '/', ppAppend)
	}
	return
}

// UniqueSlice will remove all duplicates from the given slice of strings
func UniqueSlice(a []string) (unique []string) {
	for _, aa := range a {
		if !StringInSlice(aa, unique) {
			unique = append(unique, aa)
		}
	}
	return
}

// SliceUnion will create a slice of string that contains all strings part of the passed slices
func SliceUnion(a ...[]string) []string {
	res := []string{}
	for _, aa := range a {
		res = append(res, aa...)
	}
	return UniqueSlice(res)
}

// GetTimeIn adds the passed number of seconds to the current time
func GetTimeIn(seconds int64) time.Time {
	return time.Now().Add(time.Duration(seconds) * time.Second)
}

// CompareNullableIntsWithNilAsInfinity compare two *int64 and handles nil as infinity. It returns 0 if both are equal,
// a positive value if a is greater than b, a negative value is a is less than b
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

// NewInt64 creates a new *int64
func NewInt64(i int64) *int64 {
	return &i
}

// NewInt creates a new *int
func NewInt(i int) *int {
	return &i
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

// MinInt returns the smallest of the passed integers
func MinInt(a int, ints ...int) int {
	min := a
	for _, i := range ints {
		if i < min {
			min = i
		}
	}
	return min
}

// MinInt64 returns the smallest of the passed integers
func MinInt64(a int64, ints ...int64) int64 {
	min := a
	for _, i := range ints {
		if i < min {
			min = i
		}
	}
	return min
}

// IsJWT checks if a string is a jwt
func IsJWT(token string) bool {
	arr := strings.Split(token, ".")
	if len(arr) < 3 {
		return false
	}
	for i, segment := range arr {
		if segment != "" || i < 2 { // first two segments must not be empty
			if _, err := base64.URLEncoding.DecodeString(arr[2]); err != nil {
				return false
			}
		}
	}
	return true
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
