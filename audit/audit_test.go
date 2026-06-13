package audit

import (
	"testing"

	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/contextx"
)

func TestRedactMapSensitiveKeys(t *testing.T) {
	opts := RedactOptions{MaskPhone: true, MaskLoginID: true}
	m := map[string]any{
		"name":     "Alice",
		"password": "secret123",
		"token":    "tok",
		"phone":    "13812345678",
	}
	out := RedactMap(m, opts)
	if _, ok := out["password"]; ok {
		t.Fatal("password should be removed")
	}
	if _, ok := out["token"]; ok {
		t.Fatal("token should be removed")
	}
	if out["name"] != "Alice" {
		t.Fatalf("name: %v", out["name"])
	}
	if out["phone"] != "138****5678" {
		t.Fatalf("phone: %v", out["phone"])
	}
}

func TestRedactMapNested(t *testing.T) {
	opts := RedactOptions{}
	m := map[string]any{
		"profile": map[string]any{"secret": "x"},
	}
	out := RedactMap(m, opts)
	profile, ok := out["profile"].(map[string]any)
	if !ok {
		t.Fatal("expected nested map")
	}
	if _, ok := profile["secret"]; ok {
		t.Fatal("nested secret should be removed")
	}
}

func TestMaskLoginID(t *testing.T) {
	opts := RedactOptions{MaskLoginID: true}
	m := map[string]any{"login_id": "admin@demo.com"}
	out := RedactMap(m, opts)
	got, _ := out["login_id"].(string)
	if got == "admin@demo.com" {
		t.Fatalf("login_id not masked: %s", got)
	}
}

func TestBuildChangeSummaryStruct(t *testing.T) {
	type entity struct {
		CustomerName string `json:"customerName"`
		Status       int    `json:"status"`
		Password     string `json:"password"`
	}
	before := entity{CustomerName: "A", Status: 1, Password: "p"}
	after := entity{CustomerName: "B", Status: 2, Password: "q"}
	opts := SummaryOptions{Redact: DefaultRedactOptions(RedactConfig{MaskPhone: true})}
	b, a, err := BuildChangeSummary(before, after, []string{"customerName", "status", "password"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if b["customerName"] != "A" || a["customerName"] != "B" {
		t.Fatalf("names: before=%v after=%v", b, a)
	}
	if _, ok := b["password"]; ok {
		t.Fatal("password should be excluded")
	}
}

func TestNopRecorder(t *testing.T) {
	var r Recorder = NopRecorder{}
	if err := r.Record(nil, contextx.OperatorContext{UserID: "u1"}, Entry{ModuleCode: "m"}); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultRedactOptionsFromConfig(t *testing.T) {
	opts := DefaultRedactOptions(RedactConfig{
		MaskPhone: true, MaskLoginID: false, SensitiveKeys: []string{"apiKey"},
	})
	m := map[string]any{"apiKey": "k", "phone": "13900001111"}
	out := RedactMap(m, opts)
	if _, ok := out["apiKey"]; ok {
		t.Fatal("apiKey should be sensitive")
	}
	if out["phone"] != "139****1111" {
		t.Fatalf("phone: %v", out["phone"])
	}
}

func TestAuditLoggerDisabled(t *testing.T) {
	l := New(config.AuditCfg{Enabled: false})
	l.Record(nil, Record{Operator: "u", Action: "a"})
}
