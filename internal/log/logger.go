package log

import (
	"context"
	golog "log"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/bryanl/rv-node-gen/internal/util"
)

const (
	contextKeyLogger util.ContextKey = "logger"
)

// Logger is the node gen logger.
type Logger struct {
	zapLog *zap.Logger
	logr.Logger
}

// New creates a new instance of Logger.
func New() *Logger {
	zapLog, _ := zap.NewDevelopment()
	logger := zapr.NewLogger(zapLog)

	l := &Logger{
		zapLog: zapLog,
		Logger: logger,
	}
	return l
}

// Sync syncs the logger.
func (l *Logger) Sync() {
	_ = l.zapLog.Sync()
}

// StdLogger returns a go compatible logger.
func (l *Logger) StdLogger() *golog.Logger {
	errorLog, _ := zap.NewStdLogAt(l.zapLog, zap.ErrorLevel)
	return errorLog
}

// With returns a context with a logger.
func With(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, logger)
}

// From extracts a logger from a context.
func From(ctx context.Context) *Logger {
	logger, ok := ctx.Value(contextKeyLogger).(*Logger)
	if !ok {
		return New()
	}

	return logger
}
