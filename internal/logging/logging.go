package logging

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger interface for our logging needs
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

// LogrusLogger wraps logrus.Logger to implement our Logger interface
type LogrusLogger struct {
	*logrus.Logger
}

// Setup configures and returns a logger instance
func Setup(logLevel, logFile, logFormat string, verbosity int) (Logger, error) {
	logger := logrus.New()

	// Determine log level
	level := determineLogLevel(logLevel, verbosity)
	logger.SetLevel(level)

	// Setup output
	var output io.Writer = os.Stderr
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		output = file
	}
	logger.SetOutput(output)

	// Setup formatter
	if logFormat == "json" {
		logger.SetFormatter(&JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	return &LogrusLogger{Logger: logger}, nil
}

// determineLogLevel determines the log level based on explicit level and verbosity count
func determineLogLevel(logLevel string, verbosity int) logrus.Level {
	// If explicit log level is set, use it
	if logLevel != "" {
		switch strings.ToUpper(logLevel) {
		case "DEBUG":
			return logrus.DebugLevel
		case "INFO":
			return logrus.InfoLevel
		case "WARNING", "WARN":
			return logrus.WarnLevel
		case "ERROR":
			return logrus.ErrorLevel
		case "CRITICAL", "FATAL":
			return logrus.FatalLevel
		}
	}

	// Otherwise, use verbosity count (mimicking Python implementation)
	// Default is INFO level (20), each -v reduces by 10
	level := 20 - (10 * verbosity)
	if level < 10 {
		level = 10 // DEBUG level minimum
	}

	switch {
	case level <= 10:
		return logrus.DebugLevel
	case level <= 20:
		return logrus.InfoLevel
	case level <= 30:
		return logrus.WarnLevel
	case level <= 40:
		return logrus.ErrorLevel
	default:
		return logrus.FatalLevel
	}
}

// JSONFormatter is a simple JSON formatter for logrus
type JSONFormatter struct{}

// Format formats the log entry as JSON
func (f *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := map[string]interface{}{
		"time":  entry.Time.Format("2006-01-02 15:04:05"),
		"level": strings.ToUpper(entry.Level.String()),
		"name":  "sync_tools", // Mimic Python logger name
		"msg":   entry.Message,
	}

	// Add any additional fields
	for k, v := range entry.Data {
		data[k] = v
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Add newline
	jsonData = append(jsonData, '\n')
	return jsonData, nil
}