package errors

import "testing"

func TestBizError_Error(t *testing.T) {
	e := NewBiz(BizCodeInternal, "db down")
	if got := e.Error(); got != "[500001] db down" {
		t.Fatalf("Error() = %q, want [500001] db down", got)
	}
}

func TestFailFactories(t *testing.T) {
	cases := []struct {
		fn   func(string) *BizError
		code BizCode
	}{
		{Internal, BizCodeInternal},
		{ParamInvalid, BizCodeParamInvalid},
		{ParamFormat, BizCodeParamFormat},
		{AuthMissing, BizCodeAuthMissing},
		{PermissionDenied, BizCodeForbidden},
		{ServiceUnavail, BizCodeServiceUnavail},
	}
	for _, tc := range cases {
		e := tc.fn("x")
		if e.Code != tc.code {
			t.Fatalf("code = %d, want %d", e.Code, tc.code)
		}
	}
}
