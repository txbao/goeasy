package audit

import (
	"regexp"
	"strings"
)

var defaultSensitiveKeys = []string{
	"password", "passwd", "token", "secret",
	"access_key", "secret_key", "captcha", "verify_code",
}

var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

// RedactOptions 脱敏选项。
type RedactOptions struct {
	MaskPhone         bool
	MaskLoginID       bool
	ExtraSensitiveKeys []string
}

// DefaultRedactOptions 返回默认脱敏配置。
func DefaultRedactOptions(cfg RedactConfig) RedactOptions {
	return RedactOptions{
		MaskPhone:          cfg.MaskPhone,
		MaskLoginID:        cfg.MaskLoginID,
		ExtraSensitiveKeys: cfg.SensitiveKeys,
	}
}

// RedactConfig 与 config.AuditCfg 对齐的子集（避免 audit 依赖完整 config 循环）。
type RedactConfig struct {
	MaskPhone      bool
	MaskLoginID    bool
	SensitiveKeys  []string
}

// RedactMap 对 map 做敏感字段剔除与脱敏（不修改原 map）。
func RedactMap(m map[string]any, opts RedactOptions) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	sensitive := buildSensitiveSet(opts.ExtraSensitiveKeys)
	for k, v := range m {
		if isSensitiveKey(k, sensitive) {
			continue
		}
		out[k] = redactValue(k, v, opts, sensitive)
	}
	return out
}

func buildSensitiveSet(extra []string) map[string]struct{} {
	set := make(map[string]struct{}, len(defaultSensitiveKeys)+len(extra))
	for _, k := range defaultSensitiveKeys {
		set[strings.ToLower(k)] = struct{}{}
	}
	for _, k := range extra {
		k = strings.TrimSpace(k)
		if k != "" {
			set[strings.ToLower(k)] = struct{}{}
		}
	}
	return set
}

func isSensitiveKey(key string, sensitive map[string]struct{}) bool {
	_, ok := sensitive[strings.ToLower(key)]
	return ok
}

func redactValue(key string, v any, opts RedactOptions, sensitive map[string]struct{}) any {
	switch val := v.(type) {
	case map[string]any:
		return RedactMap(val, opts)
	case string:
		return redactString(key, val, opts)
	default:
		return v
	}
}

func redactString(key, val string, opts RedactOptions) string {
	lower := strings.ToLower(key)
	if opts.MaskLoginID && (lower == "loginid" || lower == "login_id" || lower == "username" || lower == "account") {
		return maskLoginID(val)
	}
	if opts.MaskPhone && phonePattern.MatchString(val) {
		return maskPhone(val)
	}
	return val
}

func maskPhone(s string) string {
	if len(s) < 7 {
		return "****"
	}
	return s[:3] + "****" + s[len(s)-4:]
}

func maskLoginID(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 2 {
		return "**"
	}
	return s[:1] + strings.Repeat("*", len(s)-2) + s[len(s)-1:]
}
