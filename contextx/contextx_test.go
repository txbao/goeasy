package contextx

import (
	"context"
	"testing"
)

func TestOperatorContextRoundTrip(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "u-100")
	ctx = WithLoginID(ctx, "admin@demo")
	ctx = WithCustomerID(ctx, 42)
	ctx = WithOrgID(ctx, "org-1")
	ctx = WithPlatformAdmin(ctx, true)
	ctx = WithRequestID(ctx, "req-abc")
	ctx = WithTraceID(ctx, "trace-xyz")
	ctx = WithClientIP(ctx, "10.0.0.1")
	ctx = WithDeviceInfo(ctx, "Mozilla/5.0")

	op := OperatorFrom(ctx)
	if op.UserID != "u-100" || op.LoginID != "admin@demo" {
		t.Fatalf("user/login: %+v", op)
	}
	if op.CustomerID != 42 || op.OrgID != "org-1" || !op.PlatformAdmin {
		t.Fatalf("tenant: %+v", op)
	}
	if op.RequestID != "req-abc" || op.TraceID != "trace-xyz" {
		t.Fatalf("ids: %+v", op)
	}
	if op.IP != "10.0.0.1" || op.DeviceInfo != "Mozilla/5.0" {
		t.Fatalf("client: %+v", op)
	}
}

func TestOperatorFromZeroDefaults(t *testing.T) {
	op := OperatorFrom(context.Background())
	if op.UserID != "" || op.LoginID != "" || op.CustomerID != 0 || op.OrgID != "" ||
		op.PlatformAdmin || op.RequestID != "" || op.TraceID != "" || op.IP != "" || op.DeviceInfo != "" {
		t.Fatalf("expected zero values, got %+v", op)
	}
	if op := OperatorFrom(nil); op.UserID != "" {
		t.Fatalf("nil ctx: %+v", op)
	}
}

func TestIndividualGettersDefault(t *testing.T) {
	ctx := context.Background()
	if UserID(ctx) != "" || CustomerID(ctx) != 0 || PlatformAdmin(ctx) ||
		LoginID(ctx) != "" || RequestID(ctx) != "" || ClientIP(ctx) != "" || DeviceInfo(ctx) != "" {
		t.Fatal("expected empty defaults")
	}
}
