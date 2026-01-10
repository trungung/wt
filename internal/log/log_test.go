package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		logFunc  func(format string, args ...interface{})
		format   string
		args     []interface{}
		contains string
		empty    bool
	}{
		{
			name:     "debug at debug level",
			level:    LevelDebug,
			logFunc:  Debugf,
			format:   "test %s",
			args:     []interface{}{"message"},
			contains: "[DEBUG] test message",
			empty:    false,
		},
		{
			name:     "debug at warn level - suppressed",
			level:    LevelWarn,
			logFunc:  Debugf,
			format:   "test %s",
			args:     []interface{}{"message"},
			contains: "",
			empty:    true,
		},
		{
			name:     "warn at warn level",
			level:    LevelWarn,
			logFunc:  Warnf,
			format:   "test %s",
			args:     []interface{}{"warning"},
			contains: "Warning: test warning",
			empty:    false,
		},
		{
			name:     "error at error level",
			level:    LevelError,
			logFunc:  Errorf,
			format:   "test %s",
			args:     []interface{}{"error"},
			contains: "Error: test error",
			empty:    false,
		},
		{
			name:     "info at info level",
			level:    LevelInfo,
			logFunc:  Infof,
			format:   "info %s",
			args:     []interface{}{"message"},
			contains: "info message",
			empty:    false,
		},
		{
			name:     "info at warn level - suppressed",
			level:    LevelWarn,
			logFunc:  Infof,
			format:   "info %s",
			args:     []interface{}{"message"},
			contains: "",
			empty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original values
			origOutput := output
			origLevel := level
			defer func() {
				output = origOutput
				level = origLevel
			}()

			var buf bytes.Buffer
			SetOutput(&buf)
			SetLevel(tt.level)

			tt.logFunc(tt.format, tt.args...)

			got := buf.String()
			if tt.empty {
				if got != "" {
					t.Errorf("expected empty output, got %q", got)
				}
			} else {
				if !strings.Contains(got, tt.contains) {
					t.Errorf("expected output to contain %q, got %q", tt.contains, got)
				}
			}
		})
	}
}

func TestSetOutput(t *testing.T) {
	origOutput := output
	defer func() { output = origOutput }()

	var buf bytes.Buffer
	SetOutput(&buf)

	SetLevel(LevelWarn)
	Warnf("test")

	if buf.Len() == 0 {
		t.Error("expected output to be written to custom writer")
	}
}

func TestSetLevel(t *testing.T) {
	origLevel := level
	defer func() { level = origLevel }()

	SetLevel(LevelError)
	if level != LevelError {
		t.Errorf("expected level to be LevelError, got %v", level)
	}

	SetLevel(LevelDebug)
	if level != LevelDebug {
		t.Errorf("expected level to be LevelDebug, got %v", level)
	}
}
