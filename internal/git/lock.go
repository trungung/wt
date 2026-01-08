package git

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

// AcquireLock attempts to acquire a file lock for the repository.
// It returns a function that must be called to release the lock.
func AcquireLock(repoRoot string, timeout time.Duration) (func() error, error) {
	lockPath := filepath.Join(repoRoot, ".git", "wt.lock")
	fileLock := flock.New(lockPath)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	locked, err := fileLock.TryLockContext(ctx, 100*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("another wt operation is in progress (failed to acquire lock at %s: %v)", lockPath, err)
	}

	if !locked {
		return nil, fmt.Errorf("another wt operation is in progress")
	}

	return fileLock.Unlock, nil
}
