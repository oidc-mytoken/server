package logger

import (
	"bytes"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// Init initializes the logger so it just prints the error message
func Init() {
	log.SetLevel(log.ErrorLevel)
	log.SetOutput(os.Stderr)
	log.SetFormatter(&errorStringFormatter{})
}

type errorStringFormatter struct{}

// Format implements the logrus.Formatter interface
func (f *errorStringFormatter) Format(entry *log.Entry) ([]byte, error) {
	b := bytes.Buffer{}
	b.WriteString(entry.Message)
	if len(entry.Data) > 0 {
		b.WriteString(":")
	}
	for k, v := range entry.Data {
		fieldMsg := fmt.Sprintf(" %s: '%+v'", k, v)
		b.WriteString(fieldMsg)
	}
	b.WriteString("\n")
	return b.Bytes(), nil
}
