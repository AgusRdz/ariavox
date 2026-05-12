// Package srbus announces events to the native screen reader bus.
// Platform implementations use build tags.
// This is a phase 6 stub — all implementations are no-ops until then.
package srbus

// Bus announces accessibility events to the OS screen reader infrastructure.
type Bus interface {
	Announce(text string) error
	Close() error
}

// NopBus is a no-op Bus used when the SR bus is disabled or not yet implemented.
type NopBus struct{}

func (NopBus) Announce(text string) error { return nil }
func (NopBus) Close() error               { return nil }

// New returns the platform-appropriate Bus.
// Phase 6 will return real implementations per OS.
func New() Bus {
	return NopBus{}
}
