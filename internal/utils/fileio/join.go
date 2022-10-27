package fileio

import (
	"path/filepath"
)

// JoinIfFirstNotEmpty only joins filepaths if the first element is not empty
func JoinIfFirstNotEmpty(elem ...string) string {
	if len(elem) > 0 && elem[0] != "" {
		return filepath.Join(elem...)
	}
	return ""
}
