//go:build windows

package pty

import (
	"fmt"
	"os/exec"
)

// WindowsPTY is a stub for the Windows ConPTY implementation.
// Phase 1 will wire github.com/UserExistsError/conpty here.
type WindowsPTY struct{}

func New() PTY { return &WindowsPTY{} }

func (p *WindowsPTY) Start(cmd *exec.Cmd) error {
	return fmt.Errorf("pty: Windows ConPTY not yet implemented (phase 1)")
}
func (p *WindowsPTY) Read(buf []byte) (int, error)  { return 0, fmt.Errorf("not implemented") }
func (p *WindowsPTY) Write(buf []byte) (int, error) { return 0, fmt.Errorf("not implemented") }
func (p *WindowsPTY) Resize(rows, cols uint16) error { return nil }
func (p *WindowsPTY) Close() error                   { return nil }
func (p *WindowsPTY) Wait() error                    { return nil }
