// Package pty provides a platform-agnostic PTY interface.
// Platform-specific implementations use build tags.
package pty

import "os/exec"

// PTY is the interface implemented per platform.
type PTY interface {
	Start(cmd *exec.Cmd) error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Resize(rows, cols uint16) error
	Close() error
	Wait() error
}
