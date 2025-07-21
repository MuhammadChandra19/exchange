package logger

import (
	"context"
	"fmt"
	"strings"

	"github.com/muhammadchandra19/exchange/pkg/errors"
	"github.com/muhammadchandra19/exchange/pkg/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Interface is an interface that wraps the Logger methods.
//
//go:generate mockgen -source log.go -destination=mock/log_mock.go -package=logger_mock
type Interface interface {
	Debug(message string, fields ...Field)
	DebugContext(ctx context.Context, message string, fields ...Field)
	Error(err error, fields ...Field)
	ErrorContext(ctx context.Context, err error, fields ...Field)
	GetZap() *zap.Logger
	Info(message string, fields ...Field)
	InfoContext(ctx context.Context, message string, fields ...Field)
	Sync() error
	Warn(message string, fields ...Field)
	WarnContext(ctx context.Context, message string, fields ...Field)
	WithFields(fields ...Field) *Logger
}

// Logger is a wrapper around zap.Logger to provide structured logging.
type Logger struct {
	logger *zap.Logger
}

// Field holds key-value to be written to log.
type Field struct {
	Key   string
	Value any
}

// Options holds configuration options for the logger.
type Options struct {
	level           Level
	outputPaths     []string
	timeKey         string
	levelKey        string
	callerTraceSkip int
}

// Level represents the severity level of the log.
type Level string

var (
	// DebugLevel is used for debug messages.
	DebugLevel Level = "debug"
	// InfoLevel is used for informational messages.
	InfoLevel Level = "info"
	// WarnLevel is used for warning messages.
	WarnLevel Level = "warn"
	// ErrorLevel is used for error messages.
	ErrorLevel Level = "error"

	messageKey string = "message"
)

func (level Level) getZapLevel() zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel // use info level as default, same as zap's default production config
	}
}

// NewLogger creates new Logger instance with configuration options.
func NewLogger(opts ...Options) (*Logger, error) {
	cfg := zap.NewProductionConfig()
	var buildOptions []zap.Option

	// apply configuration from options
	for _, opt := range opts {
		if opt.level != "" {
			cfg.Level = zap.NewAtomicLevelAt(opt.level.getZapLevel())
		}
		if opt.outputPaths != nil {
			cfg.OutputPaths = opt.outputPaths
		}
		if opt.timeKey != "" {
			cfg.EncoderConfig.TimeKey = opt.timeKey
		}
		if opt.levelKey != "" {
			cfg.EncoderConfig.LevelKey = opt.levelKey
		}
		if opt.callerTraceSkip > 0 {
			buildOptions = append(buildOptions, zap.AddCallerSkip(opt.callerTraceSkip))
		}
	}

	// change default message key `msg` to `message`
	cfg.EncoderConfig.MessageKey = messageKey

	logger, err := cfg.Build(buildOptions...)
	return &Logger{
		logger: logger,
	}, err
}

// Sync flush the buffered log entries
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// WithLoggingLevel is used to set the minimum log level that will be logged to stdout.
// If not set, it will log `info` level and above by default
func WithLoggingLevel(level Level) Options {
	return Options{
		level: level,
	}
}

// WithOutputPaths is used to set multiple output paths that will be used to write
// logs to. The special paths "stdout" and "stderr" are interpreted as
// os.Stdout and os.Stderr. When specified without a scheme, relative file
// paths also work.
func WithOutputPaths(paths []string) Options {
	return Options{
		outputPaths: paths,
	}
}

// WithTimeKey will use key as reference for log time entry. For example, if we set it to
// "timestamp" then logger will regard "timestamp" field as time field for log reference
func WithTimeKey(key string) Options {
	return Options{
		timeKey: key,
	}
}

// WithLevelKey will use key as reference for log severity entry. For example, if we set it to
// "severity" then logger will regard "severity" field as severity level field for log reference
func WithLevelKey(key string) Options {
	return Options{
		levelKey: key,
	}
}

// WithCallerTraceSkip will skip X lines from trace log
func WithCallerTraceSkip(skip int) Options {
	return Options{
		callerTraceSkip: skip,
	}
}

// GetZap returns zap.Logger instance used by log.Logger
func (l *Logger) GetZap() *zap.Logger {
	return l.logger
}

// NewField returns Field with given key and value.
func NewField(key string, value interface{}) Field {
	return Field{key, value}
}

// Info write log with severity level info
func (l *Logger) Info(message string, fields ...Field) {
	zapFields := convertFields(fields...)
	l.logger.Info(message, zapFields...)
}

// InfoContext write log with severity level info and append request id to given fields.
func (l *Logger) InfoContext(ctx context.Context, message string, fields ...Field) {
	l.Info(message, appendRequestID(ctx, fields)...)
}

// Warn write log with severity level warn
func (l *Logger) Warn(message string, fields ...Field) {
	zapFields := convertFields(fields...)
	l.logger.Warn(message, zapFields...)
}

// WarnContext write log with severity level warn and append request id to given fields.
func (l *Logger) WarnContext(ctx context.Context, message string, fields ...Field) {
	l.Warn(message, appendRequestID(ctx, fields)...)
}

// Debug Write log with severity level debug
func (l *Logger) Debug(message string, fields ...Field) {
	zapFields := convertFields(fields...)
	l.logger.Debug(message, zapFields...)
}

// DebugContext Write log with severity level debug and append request id to given fields.
func (l *Logger) DebugContext(ctx context.Context, message string, fields ...Field) {
	l.Debug(message, appendRequestID(ctx, fields)...)
}

// Error write log with severity level error
func (l *Logger) Error(err error, fields ...Field) {
	zapFields := convertFields(fields...)
	stacktrace := ""

	if errTracer, ok := err.(errors.StackTracer); ok {
		stacktrace = strings.TrimSpace(fmt.Sprintf("%+v", errTracer.StackTrace()))
	}

	if ce := l.logger.Check(zapcore.ErrorLevel, err.Error()); ce != nil {
		if stacktrace != "" {
			// override stack trace
			ce.Stack = stacktrace
		}
		ce.Write(zapFields...)
	}
}

// ErrorContext write log with severity level error and append request id to given fields.
func (l *Logger) ErrorContext(ctx context.Context, err error, fields ...Field) {
	l.Error(err, appendRequestID(ctx, fields)...)
}

// WithFields returns a child logger with additional fields.
func (l *Logger) WithFields(fields ...Field) *Logger {
	zapFields := convertFields(fields...)
	childLogger := &Logger{
		logger: l.logger.With(zapFields...),
	}
	return childLogger
}

// convertFields transform fields to zap log fields
func convertFields(fields ...Field) []zapcore.Field {
	var zapFields []zapcore.Field
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}

// appendRequestID get request id from context and append it to given fields.
func appendRequestID(ctx context.Context, fields []Field) []Field {
	return append(fields, NewField("request_id", util.GetRequestID(ctx)))
}
