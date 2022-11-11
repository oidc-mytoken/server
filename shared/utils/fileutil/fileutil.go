// Package fileutil provides utility functions related to files.
package fileutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func evalSymlink(path string) string {
	if path == "" {
		return path
	}
	if path[0] == '~' {
		path = os.Getenv("HOME") + path[1:]
	}
	evalPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return evalPath
}

// FileExists checks if a given file exists.
func FileExists(path string) bool {
	if _, err := os.Stat(evalSymlink(path)); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		log.WithError(err).Error()
		return false
	}
}

// Append appends content to a file
func Append(path, content string, doNotCreateIfDoesNotExist ...bool) error {
	create := os.O_CREATE
	if len(doNotCreateIfDoesNotExist) > 0 && doNotCreateIfDoesNotExist[0] {
		create = 0
	}
	f, err := os.OpenFile(evalSymlink(path), os.O_APPEND|create|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	return f.Close()
}

// ReadFile reads a given file and returns the content.
func ReadFile(filename string) ([]byte, error) {
	filename = evalSymlink(filename)
	log.WithField("filepath", filename).Trace("Reading file...")
	return os.ReadFile(filename)
}

// MustReadFile reads a given file and returns the content. If an error occurs mytoken terminates.
func MustReadFile(filename string) []byte {
	file, err := ReadFile(filename)
	if err != nil {
		log.WithError(err).Fatal("Error reading config file")
	}
	log.WithField("filepath", filename).Info("Read config file")
	return file
}

// MustReadConfigFile checks if a file exists in one of the configuration
// directories and returns the content. If no file is found, mytoken exists.
func MustReadConfigFile(filename string, locations []string) ([]byte, string) {
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

// ReadConfigFile checks if a file exists in one of the configuration
// directories and returns the content. If no file is found, mytoken returns an error
func ReadConfigFile(filename string, locations []string) ([]byte, string, error) {
	for _, dir := range locations {
		if strings.HasPrefix(dir, "~") {
			homeDir := os.Getenv("HOME")
			dir = filepath.Join(homeDir, dir[1:])
		}
		filep := filepath.Join(dir, filename)
		log.WithField("filepath", filep).Debug("Looking for config file")
		if FileExists(filep) {
			data, err := ReadFile(filep)
			return data, dir, err
		}
	}
	errMsg := "Could not find config file"
	if len(locations) > 1 {
		errMsg += " in any of the possible directories"
	}
	return nil, "", errors.New(errMsg)
}
