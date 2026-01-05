package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DefaultBranch            string   `json:"defaultBranch"`
	WorktreePathTemplate     string   `json:"worktreePathTemplate"`
	WorktreeCopyPatterns     []string `json:"worktreeCopyPatterns"`
	PostCreateCmd            []string `json:"postCreateCmd"`
	DeleteBranchWithWorktree bool     `json:"deleteBranchWithWorktree"`
}

func GetConfigPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".wt.config.json")
}

func LoadConfig(repoRoot string) (*Config, error) {
	configPath := GetConfigPath(repoRoot)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Write(repoRoot string) error {
	configPath := GetConfigPath(repoRoot)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: temp file + rename
	tempFile := configPath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tempFile, configPath); err != nil {
		os.Remove(tempFile)
		return err
	}

	return nil
}

func (c *Config) GetWorktreeBase(repoRoot string) string {
	if c.WorktreePathTemplate == "" {
		return repoRoot + ".wt"
	}
	return strings.ReplaceAll(c.WorktreePathTemplate, "$REPO_PATH", repoRoot)
}
