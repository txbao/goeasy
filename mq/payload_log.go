package mq

import (
	"encoding/json"

	"github.com/txbao/goeasy/config"
)

// PayloadPreview 截断 payload 用于日志预览。
func PayloadPreview(payload json.RawMessage, maxBytes int) string {
	if maxBytes <= 0 {
		maxBytes = 512
	}
	b := []byte(payload)
	if len(b) <= maxBytes {
		return string(b)
	}
	return string(b[:maxBytes]) + "...(truncated)"
}

// PayloadLogFields 返回 MQ 消费/发布日志常用字段（不含全文 payload，除非配置开启 preview）。
func PayloadLogFields(env Envelope, cfg config.MQLogCfg) []any {
	fields := []any{
		"event_id", env.EventID,
		"event_type", env.EventType,
		"trace_id", env.TraceID,
		"payload_bytes", env.PayloadBytes(),
	}
	if cfg.LogPayload {
		max := cfg.LogPayloadMaxBytes
		if max <= 0 {
			max = 512
		}
		fields = append(fields, "payload_preview", PayloadPreview(env.Payload, max))
	}
	return fields
}
