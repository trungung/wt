package git

import (
	"reflect"
	"testing"
)

func TestParseLines(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []string
	}{
		{
			name:     "empty input",
			input:    []byte(""),
			expected: nil,
		},
		{
			name:     "single line",
			input:    []byte("foo"),
			expected: []string{"foo"},
		},
		{
			name:     "multiple lines",
			input:    []byte("foo\nbar\nbaz"),
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "with empty lines",
			input:    []byte("foo\n\nbar\n"),
			expected: []string{"foo", "bar"},
		},
		{
			name:     "with whitespace",
			input:    []byte("  foo  \n  bar  "),
			expected: []string{"foo", "bar"},
		},
		{
			name:     "only whitespace",
			input:    []byte("   \n\n   "),
			expected: nil,
		},
		{
			name:     "trailing newline",
			input:    []byte("main\nfeature\n"),
			expected: []string{"main", "feature"},
		},
		{
			name:     "mixed content",
			input:    []byte("  main  \n\n  feature/test  \n  bugfix-123  \n"),
			expected: []string{"main", "feature/test", "bugfix-123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLines(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseLines(%q) = %v, want %v", string(tt.input), got, tt.expected)
			}
		})
	}
}
