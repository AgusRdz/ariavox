//go:build darwin && !cgo

package srbus

// New returns a NopBus when CGo is disabled (e.g. cross-compilation).
// VoiceOver announcements require CGo; without it we degrade silently.
func New() Bus {
	return NopBus{}
}
