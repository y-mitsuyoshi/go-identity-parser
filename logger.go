package main

import (
	"log"
	"os"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger provides structured logging functionality
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	level := getLogLevelFromEnv()
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// getLogLevelFromEnv reads log level from environment variable
func getLogLevelFromEnv() LogLevel {
	levelStr := os.Getenv("LOG_LEVEL")
	switch levelStr {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO // Default to INFO level
	}
}

// Debug logs debug level messages
func (l *Logger) Debug(v ...interface{}) {
	if l.level <= DEBUG {
		l.logger.SetPrefix("[DEBUG] ")
		l.logger.Println(v...)
	}
}

// Debugf logs formatted debug level messages
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.logger.SetPrefix("[DEBUG] ")
		l.logger.Printf(format, v...)
	}
}

// Info logs info level messages
func (l *Logger) Info(v ...interface{}) {
	if l.level <= INFO {
		l.logger.SetPrefix("[INFO] ")
		l.logger.Println(v...)
	}
}

// Infof logs formatted info level messages
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level <= INFO {
		l.logger.SetPrefix("[INFO] ")
		l.logger.Printf(format, v...)
	}
}

// Warn logs warning level messages
func (l *Logger) Warn(v ...interface{}) {
	if l.level <= WARN {
		l.logger.SetPrefix("[WARN] ")
		l.logger.Println(v...)
	}
}

// Warnf logs formatted warning level messages
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level <= WARN {
		l.logger.SetPrefix("[WARN] ")
		l.logger.Printf(format, v...)
	}
}

// Error logs error level messages
func (l *Logger) Error(v ...interface{}) {
	if l.level <= ERROR {
		l.logger.SetPrefix("[ERROR] ")
		l.logger.Println(v...)
	}
}

// Errorf logs formatted error level messages
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.logger.SetPrefix("[ERROR] ")
		l.logger.Printf(format, v...)
	}
}

// Global logger instance
var AppLogger = NewLogger()
