//go:build darwin || linux

package pty

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// UnixPTY wraps creack/pty for macOS and Linux.
type UnixPTY struct {
	f   *os.File
	cmd *exec.Cmd
}

// NewUnix creates a PTY for unix platforms.
func New() PTY {
	return &UnixPTY{}
}

func (p *UnixPTY) Start(cmd *exec.Cmd) error {
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	// Ensure TERM is set for correct ANSI support
	cmd.Env = appendOrReplace(cmd.Env, "TERM=xterm-256color")

	f, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("pty: start failed: %w", err)
	}
	p.f = f
	p.cmd = cmd
	return nil
}

func (p *UnixPTY) Read(buf []byte) (int, error)  { return p.f.Read(buf) }
func (p *UnixPTY) Write(buf []byte) (int, error) { return p.f.Write(buf) }

func (p *UnixPTY) Resize(rows, cols uint16) error {
	return pty.Setsize(p.f, &pty.Winsize{Rows: rows, Cols: cols})
}

func (p *UnixPTY) Close() error {
	if p.f != nil {
		return p.f.Close()
	}
	return nil
}

func (p *UnixPTY) Wait() error {
	if p.cmd != nil {
		return p.cmd.Wait()
	}
	return nil
}

func appendOrReplace(env []string, kv string) []string {
	key := kv[:len(kv)-len(kv[indexOf(kv, '='):])]
	for i, e := range env {
		if len(e) >= len(key) && e[:len(key)] == key {
			env[i] = kv
			return env
		}
	}
	return append(env, kv)
}

func indexOf(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return len(s)
}
