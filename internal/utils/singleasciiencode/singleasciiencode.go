package singleasciiencode

import (
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/utils"
)

const maxFlags = 6

func NewFlagEncoder() *FlagEncoder {
	return &FlagEncoder{}
}

type FlagEncoder struct {
	names     [maxFlags]string
	f         [maxFlags]bool
	nextIndex int
}

func (fe *FlagEncoder) Set(name string, value bool) bool {
	for i, n := range fe.names {
		if i >= fe.nextIndex {
			break
		}
		if n == name {
			fe.f[i] = value
			return true
		}
	}
	if fe.nextIndex >= maxFlags {
		return false
	}
	fe.names[fe.nextIndex] = name
	fe.f[fe.nextIndex] = value
	fe.nextIndex++
	return true
}

func (fe *FlagEncoder) Sets(values map[string]bool) bool {
	names := []string{}
	for n := range values {
		names = append(names, n)
	}
	commonNames := len(utils.IntersectSlices(names, fe.names[:]))
	if maxFlags-fe.nextIndex < len(values)-commonNames { // If we cannot set all values, abort
		return false
	}
	for n, v := range values {
		if set := fe.Set(n, v); !set {
			return false
		}
	}
	return true
}

func (fe FlagEncoder) Get(name string) (value bool, found bool) {
	var index int
	var n string
	for index, n = range fe.names {
		if n == name {
			found = true
			break
		}
	}
	value = fe.f[index]
	return
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789$#"

func (fe FlagEncoder) Encode() byte {
	var flags byte
	for i, v := range fe.f {
		if i < fe.nextIndex && v {
			switch i {
			case 0:
				flags |= 0x01
			case 1:
				flags |= 0x02
			case 2:
				flags |= 0x04
			case 3:
				flags |= 0x08
			case 4:
				flags |= 0x10
			case 5:
				flags |= 0x20
			default:
				log.Error("Out of range")
			}
		}
	}
	return EncodeNumber64(flags)
}

func Decode(flags byte, names ...string) *FlagEncoder {
	fe := &FlagEncoder{
		nextIndex: len(names),
	}
	for i, n := range names {
		fe.names[i] = n
	}
	f, _ := DecodeNumber64(flags)

	if f&0x01 > 0 {
		fe.f[0] = true
	}
	if f&0x02 > 0 {
		fe.f[1] = true
	}
	if f&0x04 > 0 {
		fe.f[2] = true
	}
	if f&0x08 > 0 {
		fe.f[3] = true
	}
	if f&0x10 > 0 {
		fe.f[4] = true
	}
	if f&0x20 > 0 {
		fe.f[5] = true
	}
	return fe
}

func EncodeNumber64(n byte) byte {
	return chars[n]
}

func DecodeNumber64(e byte) (byte, bool) {
	for i, c := range chars {
		if e == byte(c) {
			return byte(i), true
		}
	}
	return 0, false
}
