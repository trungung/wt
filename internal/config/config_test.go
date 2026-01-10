package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetWorktreeBase(t *testing.T) {
	tests := []struct {
		name     string
		template string
		repoRoot string
		expected string
	}{
		{
			name:     "empty template uses default",
			template: "",
			repoRoot: "/path/to/repo",
			expected: "/path/to/repo.wt",
		},
		{
			name:     "template with REPO_PATH variable",
			template: "$REPO_PATH.wt",
			repoRoot: "/path/to/repo",
			expected: "/path/to/repo.wt",
		},
		{
			name:     "custom path without variable",
			template: "/custom/worktrees",
			repoRoot: "/path/to/repo",
			expected: "/custom/worktrees",
		},
		{
			name:     "template with suffix",
			template: "$REPO_PATH-worktrees",
			repoRoot: "/repo",
			expected: "/repo-worktrees",
		},
		{
			name:     "template with prefix",
			template: "/worktrees$REPO_PATH",
			repoRoot: "/myrepo",
			expected: "/worktrees/myrepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{WorktreePathTemplate: tt.template}
			got := cfg.GetWorktreeBase(tt.repoRoot)
			if got != tt.expected {
				t.Errorf("GetWorktreeBase(%q, %q) = %q, want %q", tt.template, tt.repoRoot, got, tt.expected)
			}
		})
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/that/does/not/exist")
	if err != nil {
		t.Errorf("LoadConfig should return empty config for missing file, got error: %v", err)
		return
	}
	if cfg == nil {
		t.Error("LoadConfig should return non-nil config")
		return
	}
	if cfg.DefaultBranch != "" {
		t.Errorf("Expected empty DefaultBranch, got %q", cfg.DefaultBranch)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	// Write config file
	configPath := filepath.Join(tempDir, ".wt.config.json")
	configContent := `{
		"defaultBranch": "main",
		"worktreePathTemplate": "$REPO_PATH.wt",
		"worktreeCopyPatterns": [".env", ".npmrc"],
		"postCreateCmd": ["npm install"],
		"deleteBranchWithWorktree": true
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want %q", cfg.DefaultBranch, "main")
	}
	if cfg.WorktreePathTemplate != "$REPO_PATH.wt" {
		t.Errorf("WorktreePathTemplate = %q, want %q", cfg.WorktreePathTemplate, "$REPO_PATH.wt")
	}
	if len(cfg.WorktreeCopyPatterns) != 2 {
		t.Errorf("WorktreeCopyPatterns length = %d, want 2", len(cfg.WorktreeCopyPatterns))
	}
	if len(cfg.PostCreateCmd) != 1 {
		t.Errorf("PostCreateCmd length = %d, want 1", len(cfg.PostCreateCmd))
	}
	if !cfg.DeleteBranchWithWorktree {
		t.Error("DeleteBranchWithWorktree should be true")
	}
}

func TestConfigWrite(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-write-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	cfg := &Config{
		DefaultBranch:            "main",
		WorktreePathTemplate:     "$REPO_PATH.wt",
		WorktreeCopyPatterns:     []string{".env"},
		PostCreateCmd:            []string{"npm install"},
		DeleteBranchWithWorktree: false,
	}

	if err := cfg.Write(tempDir); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file exists
	configPath := GetConfigPath(tempDir)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Load and verify
	loaded, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("failed to load written config: %v", err)
	}

	if loaded.DefaultBranch != cfg.DefaultBranch {
		t.Errorf("loaded DefaultBranch = %q, want %q", loaded.DefaultBranch, cfg.DefaultBranch)
	}
}

func TestCheckUnknownKeys(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-unknown-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	// Write config with unknown key
	configPath := filepath.Join(tempDir, ".wt.config.json")
	configContent := `{
		"defaultBranch": "main",
		"unknownKey": "value",
		"anotherUnknown": 123
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	unknown, err := CheckUnknownKeys(tempDir)
	if err != nil {
		t.Fatalf("CheckUnknownKeys failed: %v", err)
	}

	if len(unknown) != 2 {
		t.Errorf("expected 2 unknown keys, got %d: %v", len(unknown), unknown)
	}

	// Check that known keys are not flagged
	for _, key := range unknown {
		if key == "defaultBranch" {
			t.Error("defaultBranch should not be flagged as unknown")
		}
	}
}

func TestCheckUnknownKeys_MissingFile(t *testing.T) {
	unknown, err := CheckUnknownKeys("/nonexistent/path")
	if err != nil {
		t.Errorf("CheckUnknownKeys should not error on missing file, got: %v", err)
	}
	if unknown != nil {
		t.Errorf("expected nil unknown keys for missing file, got: %v", unknown)
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath("/repo/root")
	expected := "/repo/root/.wt.config.json"
	if path != expected {
		t.Errorf("GetConfigPath = %q, want %q", path, expected)
	}
}
