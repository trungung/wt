package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DefaultBranch            string   `json:"defaultBranch,omitempty"`
	WorktreePathTemplate     string   `json:"worktreePathTemplate,omitempty"`
	WorktreeCopyPatterns     []string `json:"worktreeCopyPatterns,omitempty"`
	PostCreateCmd            []string `json:"postCreateCmd,omitempty"`
	DeleteBranchWithWorktree bool     `json:"deleteBranchWithWorktree,omitempty"`
}

func LoadConfig(repoRoot string) (*Config, error) {
	configPath := filepath.Join(repoRoot, ".wt.config.json")
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

func (c *Config) GetWorktreeBase(repoRoot string) string {
	if c.WorktreePathTemplate == "" {
		return repoRoot + ".wt"
	}
	return strings.ReplaceAll(c.WorktreePathTemplate, "$REPO_PATH", repoRoot)
}
