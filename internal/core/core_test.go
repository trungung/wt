package core

import (
	"testing"
)

func TestMapBranchToDir(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple branch",
			branch:   "main",
			expected: "main",
			wantErr:  false,
		},
		{
			name:     "feature branch with slash",
			branch:   "feature/payment",
			expected: "feature-payment",
			wantErr:  false,
		},
		{
			name:     "nested branch with multiple slashes",
			branch:   "feature/nested/branch",
			expected: "feature-nested-branch",
			wantErr:  false,
		},
		{
			name:     "branch with underscores",
			branch:   "fix_bug_123",
			expected: "fix_bug_123",
			wantErr:  false,
		},
		{
			name:     "branch with dots",
			branch:   "release.1.0",
			expected: "release.1.0",
			wantErr:  false,
		},
		{
			name:     "branch with mixed characters",
			branch:   "feature/user-auth_v2.0",
			expected: "feature-user-auth_v2.0",
			wantErr:  false,
		},
		{
			name:     "branch with space - should error",
			branch:   "branch with space",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "branch with special chars - should error",
			branch:   "branch@123",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "branch with exclamation - should error",
			branch:   "feature!test",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapBranchToDir(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapBranchToDir(%q) error = %v, wantErr %v", tt.branch, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("MapBranchToDir(%q) = %v, want %v", tt.branch, got, tt.expected)
			}
		})
	}
}
