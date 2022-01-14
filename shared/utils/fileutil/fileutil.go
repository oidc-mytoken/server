// Package fileutil provides utility functions related to files.
package fileutil

import (
	"os"
	"path/filepath"
	"strings"

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
	f, err := os.OpenFile(evalSymlink(path), os.O_APPEND|create|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// MustReadFile reads a given config file and returns the content. If an error
// occurs mytoken terminates.
func MustReadFile(filename string) []byte {
	filename = evalSymlink(filename)
	log.WithField("filepath", filename).Trace("Found file. Reading config file ...")
	file, err := os.ReadFile(filename)
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
