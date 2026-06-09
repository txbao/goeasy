package database

import (
	"context"
	"testing"

	"github.com/txbao/goeasy/config"
)

func TestOpenDisabled(t *testing.T) {
	db, err := Open(config.DB{Enabled: false})
	if err != nil || db == nil {
		t.Fatal(err)
	}
	if db.SQLX() != nil {
		t.Fatal("noop SQLX should be nil")
	}
	if err := db.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestOpenEnabledEmptyDSN(t *testing.T) {
	_, err := Open(config.DB{Enabled: true, DSN: ""})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOpenUnsupportedORM(t *testing.T) {
	_, err := Open(config.DB{Enabled: true, DSN: "x", ORM: "gorm"})
	if err == nil {
		t.Fatal("expected error for gorm")
	}
}

func TestValidateDriver(t *testing.T) {
	if err := Validate(config.DB{Enabled: true, DSN: "postgres://x", Driver: "sqlite"}); err == nil {
		t.Fatal("expected unsupported driver error")
	}
}
