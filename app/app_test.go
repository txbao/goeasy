package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/txbao/goeasy/config"
)

func TestInitInfraDatabaseInvalid(t *testing.T) {
	cfg := &config.Config{
		AppName: "test",
		Env:     "dev",
		HTTP:    config.HTTP{Port: 18080},
		DB:      config.DB{Enabled: true, DSN: ""},
	}
	a := New(cfg)
	if err := a.InitInfra(); err == nil {
		t.Fatal("expected database error")
	}
}

func TestInitInfraDisabled(t *testing.T) {
	cfg := &config.Config{
		AppName: "test",
		Env:     "dev",
		HTTP:    config.HTTP{Port: 18081},
	}
	a := New(cfg)
	if err := a.InitInfra(); err != nil {
		t.Fatal(err)
	}
	if a.DB == nil || a.Cache == nil || a.MQ == nil {
		t.Fatal("expected noop infra")
	}
}

func TestMustLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "app_name: x\ndatabase:\n  enabled: true\n  dsn: \"\"\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.MustLoad(path)
	a := New(cfg)
	if err := a.InitInfra(); err == nil {
		t.Fatal("expected error for empty dsn")
	}
}
