package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

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
