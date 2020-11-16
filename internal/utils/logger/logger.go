package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/zachmann/mytoken/internal/config"

	"github.com/kjk/dailyrotate"
)

const accessFileFormat = "access_2006-01-02.log"

func mustGetRotateWriter(pathFormat string) io.Writer {
	w, err := dailyrotate.NewFile(pathFormat, nil)
	if err != nil {
		fmt.Printf("ERROR: Could not open %s - %s\n", time.Now().Format(pathFormat), err.Error())
		os.Exit(1)
	}
	defer w.Close()
	return w
}

func MustGetAccessLogger() io.Writer {
	var loggers []io.Writer
	if config.Get().Logging.Access.StdErr {
		loggers = append(loggers, os.Stderr)
	}
	if logDir := config.Get().Logging.Access.Dir; logDir != "" {
		loggers = append(loggers, mustGetRotateWriter(filepath.Join(logDir, accessFileFormat)))
	}
	switch len(loggers) {
	case 0:
		return nil
	case 1:
		return loggers[0]
	default:
		return io.MultiWriter(loggers...)
	}
}
