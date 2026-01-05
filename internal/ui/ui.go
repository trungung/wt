package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/yarlson/tap"
)

// Prompt uses tap.Text to get user input.
func Prompt(label string, defaultVal string) string {
	msg := label
	if defaultVal != "" {
		msg = fmt.Sprintf("%s (default: %s)", label, defaultVal)
	}
	return tap.Text(context.Background(), tap.TextOptions{
		Message:      msg,
		DefaultValue: defaultVal,
	})
}

// PromptBool uses tap.Confirm to get a boolean input.
func PromptBool(label string, defaultVal bool) bool {
	return tap.Confirm(context.Background(), tap.ConfirmOptions{
		Message:      label,
		InitialValue: defaultVal,
	})
}

// PromptSelect uses tap.Select to pick from a list of options.
func PromptSelect[T any](label string, options []tap.SelectOption[T]) T {
	return tap.Select(context.Background(), tap.SelectOptions[T]{
		Message: label,
		Options: options,
	})
}

// PromptAutocomplete uses tap.Autocomplete to get user input with suggestions.
func PromptAutocomplete(label string, suggest func(string) []string) string {
	return tap.Autocomplete(context.Background(), tap.AutocompleteOptions{
		Message:     label,
		Placeholder: "Start typing to filter...",
		Suggest:     suggest,
		MaxResults:  10,
	})
}

// PromptList uses tap.Text and splits by comma to return a list.
func PromptList(label string) []string {
	input := tap.Text(context.Background(), tap.TextOptions{
		Message: label + " (comma separated)",
	})

	if strings.TrimSpace(input) == "" {
		return nil
	}

	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
