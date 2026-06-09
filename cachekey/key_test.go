package cachekey

import "testing"

func TestEntityKey(t *testing.T) {
	got := EntityKey("mysvc", "sys_roles", "1")
	want := "mysvc:sys_roles:id:1"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
