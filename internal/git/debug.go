package git

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func debugLog(args []string, duration time.Duration) {
	if os.Getenv("WT_DEBUG") != "1" {
		return
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] git %s (took %v)\n", strings.Join(args, " "), duration)
}
