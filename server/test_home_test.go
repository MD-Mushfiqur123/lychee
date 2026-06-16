package server

import (
	"testing"

	"github.com/MD-Mushfiqur123/lychee/envconfig"
)

func setTestHome(t *testing.T, home string) {
	t.Helper()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("LYCHEE_MODELS", "")
	envconfig.ReloadServerConfig()
}
