package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("app_name: test\nhttp:\n  port: 9090\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppName != "test" || cfg.HTTP.Port != 9090 {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}
