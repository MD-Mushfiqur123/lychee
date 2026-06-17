package version

import (
	"strings"
	"testing"
)

func TestInfo(t *testing.T) {
	info := Info()

	if !strings.Contains(info, "Lychee") {
		t.Errorf("Info() = %q, want it to contain 'Lychee'", info)
	}

	if !strings.Contains(info, Version) {
		t.Errorf("Info() = %q, want it to contain Version %q", info, Version)
	}
}

func TestDefaults(t *testing.T) {
	// Default values are set at compile time, not via ldflags
	if Version == "" {
		t.Error("Version should not be empty")
	}

	if Commit == "" {
		t.Error("Commit should not be empty")
	}

	if BuildDate == "" {
		t.Error("BuildDate should not be empty")
	}
}
