package mq

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const DefaultEventVersion = "v1"

// Envelope 统一事件消息信封（NSQ payload）。
type Envelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	EventVersion  string          `json:"event_version"`
	SourceService string          `json:"source_service"`
	OccurredAt    time.Time       `json:"occurred_at"`
	TraceID       string          `json:"trace_id"`
	Payload       json.RawMessage `json:"payload"`
}

// EmptyPayload 返回空 JSON 对象 payload。
func EmptyPayload() json.RawMessage {
	return json.RawMessage("{}")
}

// NormalizePayload 将 nil/空 payload 规范为 {}，并校验 JSON 合法性。
func NormalizePayload(payload json.RawMessage) (json.RawMessage, error) {
	if len(payload) == 0 {
		return EmptyPayload(), nil
	}
	if !json.Valid(payload) {
		return nil, ErrInvalidPayload
	}
	return payload, nil
}

// NewEnvelope 构造标准事件信封。
func NewEnvelope(eventType, sourceService, traceID string, payload json.RawMessage) (Envelope, error) {
	normalized, err := NormalizePayload(payload)
	if err != nil {
		return Envelope{}, err
	}
	if traceID == "" {
		traceID = uuid.NewString()
	}
	return Envelope{
		EventID:       uuid.NewString(),
		EventType:     eventType,
		EventVersion:  DefaultEventVersion,
		SourceService: sourceService,
		OccurredAt:    time.Now().UTC(),
		TraceID:       traceID,
		Payload:       normalized,
	}, nil
}

// PayloadBytes 返回 payload 字节长度。
func (e Envelope) PayloadBytes() int {
	return len(e.Payload)
}

// PayloadInto 将 payload 解码到业务结构体。
func (e Envelope) PayloadInto(v any) error {
	if v == nil {
		return ErrInvalidPayload
	}
	if len(e.Payload) == 0 {
		return json.Unmarshal(EmptyPayload(), v)
	}
	return json.Unmarshal(e.Payload, v)
}

// Marshal 序列化为 JSON。
func (e Envelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// UnmarshalEnvelope 反序列化 JSON 信封。
func UnmarshalEnvelope(body []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return Envelope{}, err
	}
	normalized, err := NormalizePayload(env.Payload)
	if err != nil {
		return Envelope{}, err
	}
	env.Payload = normalized
	return env, nil
}
