package logger

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"

	"github.com/oidc-mytoken/server/internal/config"
)

func mustGetFile(path string) io.Writer {
	file, err := getFile(path)
	if err != nil {
		panic(err)
	}
	return file
}

func getFile(path string) (io.Writer, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	return f, errors.WithStack(err)
}

var accessLogger *exchangeableWriter

// MustGetAccessLogger open the server access logger; on failure the program exits
func MustGetAccessLogger() io.Writer {
	accessLogger = &exchangeableWriter{
		Writer: mustGetAccessLogger(),
	}
	return accessLogger
}

// MustUpdateAccessLogger updates the writer of the access logger
func MustUpdateAccessLogger() {
	accessLogger.SetOutput(mustGetAccessLogger())
}

func mustGetAccessLogger() io.Writer {
	return mustGetLogWriter(config.Get().Logging.Access, "access.log")
}

type exchangeableWriter struct {
	io.Writer
}

// SetOutput updates the internal writer
func (w *exchangeableWriter) SetOutput(out io.Writer) {
	w.Writer = out
}

func mustGetLogWriter(logConf config.LoggerConf, logfileName string) io.Writer {
	var loggers []io.Writer
	if logConf.StdErr {
		loggers = append(loggers, os.Stderr)
	}
	if logDir := logConf.Dir; logDir != "" {
		loggers = append(loggers, mustGetFile(filepath.Join(logDir, logfileName)))
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

func parseLogLevel() log.Level {
	logLevel := config.Get().Logging.Internal.Level
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithField("level", logLevel).WithError(err).Error("Unknown log level")
		return log.InfoLevel
	}
	return level
}

// Init initializes the logger
func Init() {
	log.SetLevel(log.TraceLevel) // This is not the log level for logs, this just asserts that hooks with all levels can
	// be triggered

	log.SetFormatter(
		&log.TextFormatter{
			DisableColors: true,
			ForceQuote:    true,
			FullTimestamp: true,
		},
	)
	SetOutput()
}

// SetOutput sets the logging output
func SetOutput() {
	logLevel := parseLogLevel()
	log.SetReportCaller(log.DebugLevel <= logLevel)
	log.StandardLogger().Hooks = make(log.LevelHooks)
	log.AddHook(
		&writer.Hook{
			Writer:    mustGetLogWriter(config.Get().Logging.Internal.LoggerConf, "mytoken.log"),
			LogLevels: minLogLevelToLevels(logLevel),
		},
	)
	log.SetOutput(ioutil.Discard)
}

func minLogLevelToLevels(minLevel log.Level) (levels []log.Level) {
	for l := log.PanicLevel; l <= minLevel; l++ {
		levels = append(levels, l)
	}
	return
}
