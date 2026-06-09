package mq

import (
	"encoding/json"
	"testing"

	"github.com/txbao/goeasy/config"
)

func TestNormalizePayloadEmpty(t *testing.T) {
	got, err := NormalizePayload(nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "{}" {
		t.Fatalf("got %s", got)
	}
}

func TestNewEnvelopeRoundTrip(t *testing.T) {
	payload := json.RawMessage(`{"order_id":"1"}`)
	env, err := NewEnvelope("order.created", "test", "trace-1", payload)
	if err != nil {
		t.Fatal(err)
	}
	body, err := env.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := UnmarshalEnvelope(body)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.EventType != "order.created" {
		t.Fatalf("event_type=%s", decoded.EventType)
	}
	if string(decoded.Payload) != string(payload) {
		t.Fatalf("payload=%s", decoded.Payload)
	}
}

func TestPayloadLogFields(t *testing.T) {
	env, _ := NewEnvelope("demo", "svc", "t1", json.RawMessage(`{"a":1}`))
	fields := PayloadLogFields(env, config.MQLogCfg{LogPayload: false, LogPayloadMaxBytes: 512})
	if len(fields) < 8 {
		t.Fatalf("fields=%v", fields)
	}
}
