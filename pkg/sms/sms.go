// Package sms provides a pluggable SMS sender with a noop default.
package sms

import (
	"context"
	"log"
)

// Sender delivers an SMS message.
type Sender interface {
	Send(ctx context.Context, to, message string) error
}

// NoopSender accepts messages without delivering them; it is the safe
// default when no real provider is configured.
type NoopSender struct{}

// Send records the message to the log and reports success.
func (NoopSender) Send(_ context.Context, to, message string) error {
	log.Printf("sms: noop delivery to=%q message=%q", to, message)
	return nil
}

// Config selects an SMS provider and supplies provider-specific settings.
type Config struct {
	Provider string
	Settings map[string]string
}

// New returns the configured Sender. Unknown or empty provider names fall
// back to a NoopSender so the API remains operational without vendor config.
func New(cfg Config) Sender {
	switch cfg.Provider {
	case "", "noop":
		return NoopSender{}
	default:
		log.Printf("sms: no implementation for provider %q; using noop", cfg.Provider)
		return NoopSender{}
	}
}
