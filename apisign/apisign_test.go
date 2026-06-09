package apisign

import (
	"strings"
	"testing"
)

func TestCanonicalQuery(t *testing.T) {
	got := CanonicalQuery("pageSize=10&currentPage=1&headName=中信")
	want := "currentPage=1&headName=%E4%B8%AD%E4%BF%A1&pageSize=10"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildSignString(t *testing.T) {
	s := BuildSignString("POST", "/open/v1/ping", "1700000000000", "abc", `{"traceId":"demo"}`)
	if !strings.Contains(s, "POST\n/open/v1/ping\n") {
		t.Fatalf("unexpected sign string: %q", s)
	}
}
