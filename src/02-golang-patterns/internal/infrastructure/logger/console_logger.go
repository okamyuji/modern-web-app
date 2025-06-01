package logger

import (
	"fmt"
	"log"
	"time"
)

// ConsoleLogger implements Logger using console output
type ConsoleLogger struct{}

// NewConsoleLogger creates a new ConsoleLogger
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

// Info logs an info message
func (l *ConsoleLogger) Info(msg string, fields ...interface{}) {
	l.logWithLevel("INFO", msg, fields...)
}

// Error logs an error message
func (l *ConsoleLogger) Error(msg string, fields ...interface{}) {
	l.logWithLevel("ERROR", msg, fields...)
}

// Debug logs a debug message
func (l *ConsoleLogger) Debug(msg string, fields ...interface{}) {
	l.logWithLevel("DEBUG", msg, fields...)
}

func (l *ConsoleLogger) logWithLevel(level, msg string, fields ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMsg := fmt.Sprintf("[%s] %s - %s", level, timestamp, msg)
	
	if len(fields) > 0 {
		logMsg += " "
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				logMsg += fmt.Sprintf("%v=%v ", fields[i], fields[i+1])
			}
		}
	}
	
	log.Println(logMsg)
}