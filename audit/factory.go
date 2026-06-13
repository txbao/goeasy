package audit

import "github.com/txbao/goeasy/config"

// RedactOptionsFromCfg 从 observability.audit 配置构建脱敏选项。
func RedactOptionsFromCfg(cfg config.AuditCfg) RedactOptions {
	return DefaultRedactOptions(RedactConfig{
		MaskPhone:     cfg.MaskPhoneEnabled(),
		MaskLoginID:   cfg.MaskLoginIDEnabled(),
		SensitiveKeys: cfg.SensitiveKeys,
	})
}

// BuildRecorder 根据配置组装业务 Recorder（默认 Nop；可选 Async 包装）。
func BuildRecorder(cfg config.AuditCfg, inner Recorder) Recorder {
	if inner == nil {
		inner = NopRecorder{}
	}
	if cfg.AsyncEnabled {
		return NewAsyncRecorder(inner, cfg.BufferSize)
	}
	return inner
}
