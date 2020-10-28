// Package fileutil provides utility functions related to files.
package fileutil

import (
	"io/ioutil"
	"log"
	"os"
)

// FileExists checks if a given file exists.
func FileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		log.Printf("os.Stat: %s", err.Error())
		return false
	}
}

// MustReadFile reads a given config file and returns the content. If an error
// occurs mytoken terminates.
func MustReadFile(filename string) []byte {
	log.Printf("Found %s. Reading config file ...", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading config file: %s", err.Error())
		os.Exit(1)
	}
	log.Printf("Read config file %s\n", filename)
	return file
}
