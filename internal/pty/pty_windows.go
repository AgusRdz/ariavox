//go:build windows

package pty

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/UserExistsError/conpty"
)

// WindowsPTY wraps UserExistsError/conpty for Windows ConPTY.
type WindowsPTY struct {
	cpty *conpty.ConPty
	cmd  *exec.Cmd
}

func New() PTY { return &WindowsPTY{} }

func (p *WindowsPTY) Start(cmd *exec.Cmd) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("pty: no command specified")
	}
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}

	cpty, err := conpty.Start(cmd.Path, conpty.ConPtyDimensions(80, 24))
	if err != nil {
		return fmt.Errorf("pty: conpty start failed: %w", err)
	}
	p.cpty = cpty
	p.cmd = cmd
	return nil
}

func (p *WindowsPTY) Read(buf []byte) (int, error)  { return p.cpty.Read(buf) }
func (p *WindowsPTY) Write(buf []byte) (int, error) { return p.cpty.Write(buf) }

func (p *WindowsPTY) Resize(rows, cols uint16) error {
	return p.cpty.Resize(int(cols), int(rows))
}

func (p *WindowsPTY) Close() error {
	if p.cpty != nil {
		return p.cpty.Close()
	}
	return nil
}

func (p *WindowsPTY) Wait() error {
	if p.cpty != nil {
		_, err := p.cpty.Wait(nil)
		return err
	}
	return nil
}
