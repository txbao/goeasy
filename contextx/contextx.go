package contextx

import "context"

type key int

const (
	keyUserID key = iota + 1
	keyOrgID
	keyTraceID
	keyCustomerID
	keyPlatformAdmin
	keyLoginID
	keyRequestID
	keyClientIP
	keyDeviceInfo
)

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyUserID, id)
}

func UserID(ctx context.Context) string {
	if v, ok := ctx.Value(keyUserID).(string); ok {
		return v
	}
	return ""
}

func WithOrgID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyOrgID, id)
}

func OrgID(ctx context.Context) string {
	if v, ok := ctx.Value(keyOrgID).(string); ok {
		return v
	}
	return ""
}

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyTraceID, id)
}

func TraceID(ctx context.Context) string {
	if v, ok := ctx.Value(keyTraceID).(string); ok {
		return v
	}
	return ""
}

func WithCustomerID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, keyCustomerID, id)
}

func CustomerID(ctx context.Context) int64 {
	if v, ok := ctx.Value(keyCustomerID).(int64); ok {
		return v
	}
	return 0
}

func WithPlatformAdmin(ctx context.Context, admin bool) context.Context {
	return context.WithValue(ctx, keyPlatformAdmin, admin)
}

func PlatformAdmin(ctx context.Context) bool {
	if v, ok := ctx.Value(keyPlatformAdmin).(bool); ok {
		return v
	}
	return false
}

func WithLoginID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyLoginID, id)
}

func LoginID(ctx context.Context) string {
	if v, ok := ctx.Value(keyLoginID).(string); ok {
		return v
	}
	return ""
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRequestID, id)
}

func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(keyRequestID).(string); ok {
		return v
	}
	return ""
}

func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, keyClientIP, ip)
}

func ClientIP(ctx context.Context) string {
	if v, ok := ctx.Value(keyClientIP).(string); ok {
		return v
	}
	return ""
}

func WithDeviceInfo(ctx context.Context, info string) context.Context {
	return context.WithValue(ctx, keyDeviceInfo, info)
}

func DeviceInfo(ctx context.Context) string {
	if v, ok := ctx.Value(keyDeviceInfo).(string); ok {
		return v
	}
	return ""
}

// OperatorContext 操作人上下文（从 context.Context 聚合，无 gin 依赖）。
type OperatorContext struct {
	UserID        string
	LoginID       string
	CustomerID    int64
	OrgID         string
	PlatformAdmin bool
	RequestID     string
	TraceID       string
	IP            string
	DeviceInfo    string
}

// OperatorFrom 从 context 聚合操作人信息。
func OperatorFrom(ctx context.Context) OperatorContext {
	if ctx == nil {
		return OperatorContext{}
	}
	return OperatorContext{
		UserID:        UserID(ctx),
		LoginID:       LoginID(ctx),
		CustomerID:    CustomerID(ctx),
		OrgID:         OrgID(ctx),
		PlatformAdmin: PlatformAdmin(ctx),
		RequestID:     RequestID(ctx),
		TraceID:       TraceID(ctx),
		IP:            ClientIP(ctx),
		DeviceInfo:    DeviceInfo(ctx),
	}
}
