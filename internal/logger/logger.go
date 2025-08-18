package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger wraps slog.Logger with convenience methods
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // json, text
}

// Global logger instance
var defaultLogger *Logger

// InitLogger initializes the global logger with the given configuration
func InitLogger(config Config) {
	var level slog.Level
	switch strings.ToLower(config.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	switch strings.ToLower(config.Format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	defaultLogger = &Logger{Logger: logger}
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		// Initialize with default config if not already initialized
		InitLogger(Config{Level: "info", Format: "text"})
	}
	return defaultLogger
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{Logger: l.Logger.With()}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{Logger: l.Logger.With(args...)}
}

// WithField returns a logger with a single additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Logger: l.Logger.With(key, value)}
}

// Convenience methods
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// HTTP Middleware for request logging
func GinLogger() gin.HandlerFunc {
	logger := GetLogger()

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request info
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		// Log fields
		fields := map[string]interface{}{
			"method":     method,
			"path":       path,
			"status":     statusCode,
			"latency":    latency.String(),
			"client_ip":  clientIP,
			"user_agent": userAgent,
			"body_size":  bodySize,
		}

		// Add error if exists
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// Log based on status code
		message := fmt.Sprintf("%s %s", method, path)
		loggerWithFields := logger.WithFields(fields)

		switch {
		case statusCode >= 500:
			loggerWithFields.Error(message)
		case statusCode >= 400:
			loggerWithFields.Warn(message)
		default:
			loggerWithFields.Info(message)
		}
	}
}

// Global convenience functions
func Debug(msg string, args ...interface{}) {
	GetLogger().Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	GetLogger().Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	GetLogger().Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	GetLogger().Error(msg, args...)
}

func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

func WithFields(fields map[string]interface{}) *Logger {
	return GetLogger().WithFields(fields)
}

func WithField(key string, value interface{}) *Logger {
	return GetLogger().WithField(key, value)
}
