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
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return err
	}

	if err := os.Rename(tempFile, configPath); err != nil {
		_ = os.Remove(tempFile)
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

// CheckUnknownKeys returns an error if the config file contains keys not in the Config struct.
func CheckUnknownKeys(repoRoot string) ([]string, error) {
	configPath := GetConfigPath(repoRoot)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	known := map[string]bool{
		"defaultBranch":            true,
		"worktreePathTemplate":     true,
		"worktreeCopyPatterns":     true,
		"postCreateCmd":            true,
		"deleteBranchWithWorktree": true,
	}

	var unknown []string
	for k := range raw {
		if !known[k] {
			unknown = append(unknown, k)
		}
	}
	return unknown, nil
}
