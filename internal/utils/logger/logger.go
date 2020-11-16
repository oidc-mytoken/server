package logger

import (
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/zachmann/mytoken/internal/config"
)

func mustGetFile(path string) io.Writer {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		return file
	}
	panic(err)
}

func MustGetAccessLogger() io.Writer {
	return mustGetLogWriter(config.Get().Logging.Access, "access.log")
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

func Init() {
	log.SetLevel(parseLogLevel())
	if log.IsLevelEnabled(log.DebugLevel) {
		log.SetReportCaller(true)
	}
	log.SetOutput(mustGetLogWriter(config.Get().Logging.Internal, "mytoken.log"))
}
