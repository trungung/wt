package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// PromptWithError uses huh.NewInput to get user input, returning an error on failure.
func PromptWithError(label string, defaultVal string) (string, error) {
	msg := label
	if defaultVal != "" {
		msg = fmt.Sprintf("%s (default: %s)", label, defaultVal)
	}
	var result string
	err := huh.NewInput().
		Title(msg).
		Value(&result).
		Run()
	if err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}
	if result == "" {
		result = defaultVal
	}
	return result, nil
}

// PromptRequiredWithError uses huh.NewInput to get user input and requires it to be non-empty.
func PromptRequiredWithError(label string) (string, error) {
	var result string
	err := huh.NewInput().
		Title(label).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("this field is required")
			}
			return nil
		}).
		Value(&result).
		Run()
	if err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}
	return result, nil
}

// PromptBoolWithError uses huh.NewConfirm to get a boolean input.
func PromptBoolWithError(label string, defaultVal bool) (bool, error) {
	result := defaultVal
	err := huh.NewConfirm().
		Title(label).
		Affirmative("Yes").
		Negative("No").
		Value(&result).
		Run()
	if err != nil {
		return false, fmt.Errorf("prompt failed: %w", err)
	}
	return result, nil
}

// PromptSelectWithError uses huh.NewSelect to pick from a list of options.
func PromptSelectWithError[T comparable](label string, options []huh.Option[T]) (T, error) {
	var result T
	err := huh.NewSelect[T]().
		Title(label).
		Options(options...).
		Value(&result).
		Run()
	if err != nil {
		return result, fmt.Errorf("prompt failed: %w", err)
	}
	return result, nil
}

// PromptListWithError uses huh.NewInput and splits by comma to return a list.
func PromptListWithError(label string) ([]string, error) {
	var input string
	err := huh.NewInput().
		Title(label + " (comma separated)").
		Value(&input).
		Run()
	if err != nil {
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	if strings.TrimSpace(input) == "" {
		return []string{}, nil
	}

	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result, nil
}

// Prompt uses huh.NewInput to get user input.
// Deprecated: Use PromptWithError instead to handle errors gracefully.
func Prompt(label string, defaultVal string) string {
	result, err := PromptWithError(label, defaultVal)
	if err != nil {
		panic(err)
	}
	return result
}

// PromptRequired uses huh.NewInput to get user input and requires it to be non-empty.
// Deprecated: Use PromptRequiredWithError instead to handle errors gracefully.
func PromptRequired(label string) string {
	result, err := PromptRequiredWithError(label)
	if err != nil {
		panic(err)
	}
	return result
}

// PromptBool uses huh.NewConfirm to get a boolean input.
// Deprecated: Use PromptBoolWithError instead to handle errors gracefully.
func PromptBool(label string, defaultVal bool) bool {
	result, err := PromptBoolWithError(label, defaultVal)
	if err != nil {
		panic(err)
	}
	return result
}

// PromptSelect uses huh.NewSelect to pick from a list of options.
// Deprecated: Use PromptSelectWithError instead to handle errors gracefully.
func PromptSelect[T comparable](label string, options []huh.Option[T]) T {
	result, err := PromptSelectWithError(label, options)
	if err != nil {
		panic(err)
	}
	return result
}

// PromptList uses huh.NewInput and splits by comma to return a list.
// Deprecated: Use PromptListWithError instead to handle errors gracefully.
func PromptList(label string) []string {
	result, err := PromptListWithError(label)
	if err != nil {
		panic(err)
	}
	return result
}
