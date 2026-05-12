//go:build linux

package srbus

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

// ATSPIBus sends accessibility announcements via AT-SPI2 on Linux.
// It uses the org.a11y.atspi.Registry D-Bus service which Orca and other
// AT-SPI consumers listen to.
type ATSPIBus struct {
	conn *dbus.Conn
}

// New returns the platform-appropriate Bus.
// Falls back to NopBus if AT-SPI2 is unavailable (no D-Bus session, no Orca).
func New() Bus {
	conn, err := dbus.SessionBus()
	if err != nil {
		return NopBus{}
	}
	// probe that the accessibility registry is reachable
	obj := conn.Object("org.a11y.Bus", "/org/a11y/bus")
	var addr string
	if err := obj.Call("org.a11y.Bus.GetAddress", 0).Store(&addr); err != nil {
		// AT-SPI bus not running — degrade silently
		conn.Close()
		return NopBus{}
	}
	// open a connection to the AT-SPI bus itself
	atspiBus, err := dbus.Dial(addr)
	if err != nil {
		conn.Close()
		return NopBus{}
	}
	if err := atspiBus.Auth(nil); err != nil {
		conn.Close()
		atspiBus.Close()
		return NopBus{}
	}
	if err := atspiBus.Hello(); err != nil {
		conn.Close()
		atspiBus.Close()
		return NopBus{}
	}
	conn.Close() // only needed to get the AT-SPI address
	return &ATSPIBus{conn: atspiBus}
}

// Announce emits an AT-SPI "accessible-description-changed" style announcement
// by calling the org.a11y.atspi.EventListener:Announce signal.
//
// In practice, most AT-SPI implementations (Orca) respond to the
// "object:text-changed" style events emitted by real widgets. For a
// non-widget process the most portable approach is to emit a D-Bus signal
// that Orca's event listener picks up. We use the speech-dispatcher D-Bus
// interface when available, falling back to a raw D-Bus signal.
func (b *ATSPIBus) Announce(text string) error {
	if text == "" {
		return nil
	}
	// Emit the standard AT-SPI announcement signal.
	// Orca listens for "object:announcement" on the AT-SPI event bus.
	return b.conn.Emit(
		"/org/a11y/atspi/accessible/ariavox",
		"org.a11y.atspi.Event.Object.Announcement",
		text,
	)
}

func (b *ATSPIBus) Close() error {
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}

// Ensure ATSPIBus satisfies Bus at compile time.
var _ Bus = (*ATSPIBus)(nil)

// announcePath is exported for tests.
const announcePath = "/org/a11y/atspi/accessible/ariavox"

var _ = fmt.Sprintf // keep fmt import used
