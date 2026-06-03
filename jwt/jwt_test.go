package jwt

import (
	"testing"

	"github.com/txbao/goeasy/config"
)

func TestGenerateParse(t *testing.T) {
	tok, err := New(config.JWTCfg{Enabled: true, Secret: "secret", Issuer: "test", ExpireMin: 10})
	if err != nil {
		t.Fatal(err)
	}
	s, err := tok.Generate("user-1")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := tok.Parse(s)
	if err != nil || claims.Subject != "user-1" {
		t.Fatalf("claims=%+v err=%v", claims, err)
	}
}
