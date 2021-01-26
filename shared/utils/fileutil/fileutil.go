// Package fileutil provides utility functions related to files.
package fileutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// FileExists checks if a given file exists.
func FileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		log.WithError(err).Error()
		return false
	}
}

// MustReadFile reads a given config file and returns the content. If an error
// occurs mytoken terminates.
func MustReadFile(filename string) []byte {
	log.WithField("filepath", filename).Trace("Found file. Reading config file ...")
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.WithError(err).Error("Error reading config file")
		os.Exit(1)
	}
	log.WithField("filepath", filename).Info("Read config file")
	return file
}

// ReadConfigFile checks if a file exists in one of the configuration
// directories and returns the content. If no file is found, mytoken exists.
func ReadConfigFile(filename string, locations []string) ([]byte, string) {
	for _, dir := range locations {
		if strings.HasPrefix(dir, "~") {
			homeDir := os.Getenv("HOME")
			dir = filepath.Join(homeDir, dir[1:])
		}
		filep := filepath.Join(dir, filename)
		log.WithField("filepath", filep).Debug("Looking for config file")
		if FileExists(filep) {
			return MustReadFile(filep), dir
		}
	}
	errMsg := "Could not find config file"
	if len(locations) > 1 {
		errMsg += " in any of the possible directories"
	}
	log.WithField("filepath", filename).Fatal(errMsg)
	return nil, ""
}
