// Package srbus announces events to the native screen reader bus.
// Platform implementations use build tags — never runtime.GOOS.
package srbus

// Bus announces accessibility events to the OS screen reader infrastructure.
type Bus interface {
	// Announce sends text to the screen reader. Errors are non-fatal.
	Announce(text string) error
	Close() error
}

// NopBus is a no-op Bus used when the SR bus is disabled.
type NopBus struct{}

func (NopBus) Announce(text string) error { return nil }
func (NopBus) Close() error               { return nil }
