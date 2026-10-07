package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	t.Setenv("JEV_ENV_KEEP", "existing")
	path := filepath.Join(t.TempDir(), "keys.env")
	t.Setenv("JEV_ENV_NEW", "")
	os.Unsetenv("JEV_ENV_NEW")
	t.Cleanup(func() { os.Unsetenv("JEV_ENV_NEW") })
	if err := os.WriteFile(path, []byte("\ufeff# local file\nJEV_ENV_KEEP=other\nJEV_ENV_NEW='abc=123'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := LoadEnv(path); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("JEV_ENV_KEEP") != "existing" || os.Getenv("JEV_ENV_NEW") != "abc=123" {
		t.Fatal("environment precedence/parser")
	}
}
