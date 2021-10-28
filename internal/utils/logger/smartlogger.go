package logger

import (
	"bytes"
	"io"
	"path/filepath"

	"github.com/gliderlabs/ssh"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"

	"github.com/oidc-mytoken/server/internal/config"
)

type SmartLogger struct {
	*log.Entry
	rootHook *rootHook
	ctx      smartLoggerContext
}

type smartLoggerContext struct {
	buffer *bytes.Buffer
	id     string
}

type rootHook struct {
	buffer log.Hook
	error  *errorHook
}

func (h *rootHook) Levels() []log.Level {
	return log.AllLevels
}
func (h *rootHook) Fire(e *log.Entry) error {
	if !h.error.firedBefore {
		if err := h.buffer.Fire(e); err != nil {
			return err
		}
	}
	if h.error.firedBefore || log.ErrorLevel >= e.Level {
		if err := h.error.Fire(e); err != nil {
			return err
		}
	}
	return nil
}

type errorHook struct {
	*smartLoggerContext
	firedBefore bool
	file        io.Writer
}

func (h *errorHook) Levels() []log.Level {
	return log.AllLevels // we must be triggered at
}
func (h *errorHook) Fire(e *log.Entry) (err error) {
	var logData []byte
	if h.firedBefore {
		logData, err = e.Bytes()
		if err != nil {
			return
		}
	} else {
		logData = h.smartLoggerContext.buffer.Bytes()
		// from now on we will log all future log messages directly to file (if there are any)
		h.firedBefore = true
		h.smartLoggerContext.buffer.Reset()
	}
	file, errr := h.getFile()
	if errr != nil {
		return errr
	}
	if _, err = file.Write(logData); err != nil {
		return
	}
	return
}

func (h *errorHook) getFile() (io.Writer, error) {
	var err error
	if h.file == nil {
		h.file, err = getFile(filepath.Join(config.Get().Logging.Internal.Smart.Dir, h.smartLoggerContext.id))
	}
	return h.file, err
}
func newErrorHook(ctx *smartLoggerContext) *errorHook {
	return &errorHook{
		smartLoggerContext: ctx,
	}
}
func newBufferHook(ctx *smartLoggerContext) log.Hook {
	return &writer.Hook{
		Writer:    ctx.buffer,
		LogLevels: log.AllLevels,
	}
}
func newRootHook(ctx *smartLoggerContext) *rootHook {
	return &rootHook{
		buffer: newBufferHook(ctx),
		error:  newErrorHook(ctx),
	}
}

func smartPrepareLogger(rootH *rootHook) *log.Logger {
	std := log.StandardLogger()
	logger := &log.Logger{
		Out:          std.Out,
		Hooks:        make(log.LevelHooks),
		Formatter:    std.Formatter,
		ReportCaller: std.ReportCaller,
		Level:        std.Level,
		ExitFunc:     std.ExitFunc,
	}
	for l, hs := range std.Hooks {
		logger.Hooks[l] = append([]log.Hook{}, hs...)
	}
	logger.Hooks.Add(rootH)
	return logger
}

func getLogEntry(id string, logger *log.Logger) *log.Entry {
	return logger.WithField("requestid", id)
}

func getIDlogger(id string) log.Ext1FieldLogger {
	if !config.Get().Logging.Internal.Smart.Enabled {
		return getLogEntry(id, log.StandardLogger())
	}
	smartLog := &SmartLogger{
		ctx: smartLoggerContext{
			buffer: new(bytes.Buffer),
			id:     id,
		},
	}
	smartLog.rootHook = newRootHook(&smartLog.ctx)
	logger := smartPrepareLogger(smartLog.rootHook)
	smartLog.Entry = getLogEntry(id, logger)
	return smartLog
}

// GetRequestLogger returns a logrus.Ext1FieldLogger that always includes a request's id
func GetRequestLogger(ctx *fiber.Ctx) log.Ext1FieldLogger {
	return getIDlogger(ctx.Locals("requestid").(string))
}

// GetSSHRequestLogger returns a logrus.Ext1FieldLogger that always includes an ssh request's id
func GetSSHRequestLogger(ctx ssh.Context) log.Ext1FieldLogger {
	return getIDlogger(ctx.SessionID())
}
