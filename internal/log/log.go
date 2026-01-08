package log

import (
	"fmt"
	"io"
	"os"
)

// Level represents the logging level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	output io.Writer = os.Stderr
	level  Level     = LevelWarn
)

func init() {
	if os.Getenv("WT_DEBUG") == "1" {
		level = LevelDebug
	}
}

// SetOutput sets the output writer (useful for testing)
func SetOutput(w io.Writer) {
	output = w
}

// SetLevel sets the minimum log level
func SetLevel(l Level) {
	level = l
}

// Debugf logs a debug message (only when WT_DEBUG=1)
func Debugf(format string, args ...interface{}) {
	if level <= LevelDebug {
		_, _ = fmt.Fprintf(output, "[DEBUG] "+format+"\n", args...)
	}
}

// Infof logs an info message
func Infof(format string, args ...interface{}) {
	if level <= LevelInfo {
		_, _ = fmt.Fprintf(output, format+"\n", args...)
	}
}

// Warnf logs a warning message
func Warnf(format string, args ...interface{}) {
	if level <= LevelWarn {
		_, _ = fmt.Fprintf(output, "Warning: "+format+"\n", args...)
	}
}

// Errorf logs an error message
func Errorf(format string, args ...interface{}) {
	if level <= LevelError {
		_, _ = fmt.Fprintf(output, "Error: "+format+"\n", args...)
	}
}
