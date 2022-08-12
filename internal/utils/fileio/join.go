package fileio

import (
	"path/filepath"
)

func JoinIfFirstNotEmpty(elem ...string) string {
	if len(elem) > 0 && elem[0] != "" {
		return filepath.Join(elem...)
	}
	return ""
}
