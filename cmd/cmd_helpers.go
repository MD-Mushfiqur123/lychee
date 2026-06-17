package cmd

import (
	"fmt"
	"strings"
	"time"
)

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d >= time.Hour {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

// wrapServerError wraps an error with context and checks if the Lychee server
// is not running (connection refused). If so, it returns a user-friendly message
// with the underlying error appended.
func wrapServerError(action string, err error) error {
	if isConnectionRefused(err) {
		return fmt.Errorf("Error: Could not connect to Lychee server. Is 'lychee serve' running?\n  failed to %s: %w", action, err)
	}
	return fmt.Errorf("failed to %s: %w", action, err)
}

// isConnectionRefused checks if the error is due to the Lychee server not running.
func isConnectionRefused(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no connection could be made") ||
		strings.Contains(msg, "connectex")
}
