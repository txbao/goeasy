package validator

import "testing"

type sample struct {
	Name string `json:"name" validate:"required"`
}

func TestFormatChineseRequired(t *testing.T) {
	err := Validate(sample{})
	if err == nil {
		t.Fatal("expected error")
	}
	msg := Format(err)
	if msg == "" || msg == err.Error() {
		t.Fatalf("expected zh translation, got %q", msg)
	}
}
