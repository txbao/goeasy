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
