package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/trungung/wt/internal/config"
	"github.com/trungung/wt/internal/git"
)

type HealthLevel string

const (
	LevelOK    HealthLevel = "OK"
	LevelWARN  HealthLevel = "WARN"
	LevelERROR HealthLevel = "ERROR"
)

type HealthCheck struct {
	Name    string
	Level   HealthLevel
	Message string
}

func RunHealthCheck() ([]HealthCheck, bool) {
	var checks []HealthCheck
	hasError := false

	add := func(name string, level HealthLevel, message string) {
		checks = append(checks, HealthCheck{Name: name, Level: level, Message: message})
		if level == LevelERROR {
			hasError = true
		}
	}

	// 1. Repo root
	root, err := git.GetRepoRoot()
	if err != nil {
		add("Repo root", LevelERROR, fmt.Sprintf("Not a git repository: %v", err))
		return checks, true
	}
	add("Repo root", LevelOK, root)

	// 2. Config validity
	configPath := config.GetConfigPath(root)
	var cfg *config.Config
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			add("Config", LevelERROR, fmt.Sprintf("Failed to read config: %v", err))
		} else {
			var temp map[string]interface{}
			if err := json.Unmarshal(data, &temp); err != nil {
				add("Config", LevelERROR, fmt.Sprintf("Invalid JSON: %v", err))
			} else {
				unknown, _ := config.CheckUnknownKeys(root)
				if len(unknown) > 0 {
					add("Config", LevelWARN, fmt.Sprintf("Unknown keys: %v", unknown))
				} else {
					add("Config", LevelOK, "Valid")
				}
			}
		}
		cfg, err = config.LoadConfig(root)
		if err != nil {
			// If LoadConfig fails, we use a blank config for the rest of the checks
			// to avoid panics, but the ERROR is already reported above.
			cfg = &config.Config{}
		}
	} else {
		add("Config", LevelOK, "Not present (using defaults)")
		cfg = &config.Config{}
	}

	// 3. Default branch
	defaultBranch := cfg.DefaultBranch
	if defaultBranch == "" {
		detected, err := git.GetDefaultBranch()
		if err != nil {
			add("Default branch", LevelERROR, "Could not determine default branch via origin/HEAD. Please set 'defaultBranch' in config.")
		} else {
			add("Default branch", LevelOK, detected)
			defaultBranch = detected
		}
	} else {
		add("Default branch", LevelOK, fmt.Sprintf("%s (override)", defaultBranch))
	}

	// 4. Worktree base directory writability
	wtRoot := cfg.GetWorktreeBase(root)
	// Check if wtRoot exists and is writable, OR if parent is writable
	parent := filepath.Dir(wtRoot)
	if _, err := os.Stat(wtRoot); err == nil {
		// wtRoot exists, check if writable (best effort check)
		f, err := os.Create(filepath.Join(wtRoot, ".wt.tmp"))
		if err != nil {
			add("Worktree base", LevelERROR, fmt.Sprintf("Directory %s is not writable: %v", wtRoot, err))
		} else {
			f.Close()
			os.Remove(f.Name())
			add("Worktree base", LevelOK, wtRoot)
		}
	} else {
		// wtRoot doesn't exist, check parent
		if _, err := os.Stat(parent); err == nil {
			f, err := os.Create(filepath.Join(parent, ".wt.tmp"))
			if err != nil {
				add("Worktree base", LevelERROR, fmt.Sprintf("Parent directory %s is not writable: %v", parent, err))
			} else {
				f.Close()
				os.Remove(f.Name())
				add("Worktree base", LevelOK, fmt.Sprintf("%s (to be created)", wtRoot))
			}
		} else {
			add("Worktree base", LevelERROR, fmt.Sprintf("Parent directory %s does not exist", parent))
		}
	}

	// 5. Copy patterns
	if len(cfg.WorktreeCopyPatterns) > 0 {
		for _, p := range cfg.WorktreeCopyPatterns {
			matches, err := filepath.Glob(filepath.Join(root, p))
			if err != nil {
				add("Copy patterns", LevelWARN, fmt.Sprintf("Pattern %q is invalid: %v", p, err))
			} else if len(matches) == 0 {
				add("Copy patterns", LevelWARN, fmt.Sprintf("Pattern %q matches nothing in repo", p))
			} else {
				// Check if any matched file is tracked by git
				for _, m := range matches {
					rel, _ := filepath.Rel(root, m)
					if git.IsTracked(root, rel) {
						add("Copy patterns", LevelWARN, fmt.Sprintf("File %q is tracked by git; worktreeCopyPattern is redundant for it", rel))
					}
				}
			}
		}
	}

	// 6. Collision Scan (WARN)
	branches, err := git.ListLocalBranches()
	if err == nil {
		mapping := make(map[string]string) // dir -> branch
		for _, b := range branches {
			dir, err := MapBranchToDir(b)
			if err != nil {
				continue
			}
			if other, exists := mapping[dir]; exists {
				add("Collisions", LevelWARN, fmt.Sprintf("Branches %q and %q will both map to directory %q", b, other, dir))
			}
			mapping[dir] = b
		}
	}

	return checks, hasError
}
