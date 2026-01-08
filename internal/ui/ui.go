package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// Prompt uses huh.NewInput to get user input.
func Prompt(label string, defaultVal string) string {
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
		panic(err)
	}
	if result == "" {
		result = defaultVal
	}
	return result
}

// PromptRequired uses huh.NewInput to get user input and requires it to be non-empty.
func PromptRequired(label string) string {
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
		panic(err)
	}
	return result
}

// PromptBool uses huh.NewConfirm to get a boolean input.
func PromptBool(label string, defaultVal bool) bool {
	var result bool
	err := huh.NewConfirm().
		Title(label).
		Affirmative("Yes").
		Negative("No").
		Value(&result).
		Run()
	if err != nil {
		panic(err)
	}
	return result
}

// PromptSelect uses huh.NewSelect to pick from a list of options.
func PromptSelect[T comparable](label string, options []huh.Option[T]) T {
	var result T
	err := huh.NewSelect[T]().
		Title(label).
		Options(options...).
		Value(&result).
		Run()
	if err != nil {
		panic(err)
	}
	return result
}

// PromptList uses huh.NewInput and splits by comma to return a list.
func PromptList(label string) []string {
	var input string
	err := huh.NewInput().
		Title(label + " (comma separated)").
		Value(&input).
		Run()
	if err != nil {
		panic(err)
	}

	if strings.TrimSpace(input) == "" {
		return []string{}
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
