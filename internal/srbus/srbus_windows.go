//go:build windows

package srbus

import (
	"encoding/binary"
	"fmt"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

const pipeName = `\\.\pipe\ariavox-sr`

// PipeBus writes announcements to a named pipe that screen readers
// (NVDA addons, JAWS scripts) can connect to and relay via UIA/MSAA.
type PipeBus struct {
	mu     sync.Mutex
	handle windows.Handle
}

// New returns the platform-appropriate Bus.
// Creates the named pipe; falls back to NopBus if creation fails
// (e.g. another ariavox instance is already running).
func New() Bus {
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{}))}
	h, err := windows.CreateNamedPipe(
		windows.StringToUTF16Ptr(pipeName),
		windows.PIPE_ACCESS_OUTBOUND,
		windows.PIPE_TYPE_MESSAGE|windows.PIPE_READMODE_MESSAGE|windows.PIPE_WAIT,
		1,    // max instances
		4096, // out buffer
		0,    // in buffer
		0,    // default timeout
		&sa,
	)
	if err != nil {
		return NopBus{}
	}
	return &PipeBus{handle: h}
}

// Announce writes a length-prefixed UTF-8 message to the pipe.
// Screen reader consumers read: [4-byte LE length][utf-8 text].
func (b *PipeBus) Announce(text string) error {
	if text == "" {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	payload := []byte(text)
	hdr := make([]byte, 4)
	binary.LittleEndian.PutUint32(hdr, uint32(len(payload)))
	msg := append(hdr, payload...)

	var written uint32
	if err := windows.WriteFile(b.handle, msg, &written, nil); err != nil {
		return fmt.Errorf("srbus: pipe write: %w", err)
	}
	return nil
}

func (b *PipeBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.handle != windows.InvalidHandle {
		err := windows.CloseHandle(b.handle)
		b.handle = windows.InvalidHandle
		return err
	}
	return nil
}

// Ensure PipeBus satisfies Bus at compile time.
var _ Bus = (*PipeBus)(nil)
