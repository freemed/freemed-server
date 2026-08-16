package sms

import (
	"context"
	"testing"
)

func TestNew_DefaultsToNoop(t *testing.T) {
	for _, provider := range []string{"", "noop", "twilio", "unknown"} {
		s := New(Config{Provider: provider})
		if _, ok := s.(NoopSender); !ok {
			t.Errorf("New(%q) returned %T, want NoopSender", provider, s)
		}
	}
}

func TestNoopSender_Send(t *testing.T) {
	var s NoopSender
	if err := s.Send(context.Background(), "+15551234567", "hello"); err != nil {
		t.Errorf("NoopSender.Send() unexpected error: %v", err)
	}
}
