//go:build darwin || linux

package main

import (
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/AgusRdz/ariavox/internal/pty"
)

func notifySignals(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)
}

func handleSignals(ch <-chan os.Signal, p pty.PTY) {
	for sig := range ch {
		switch sig {
		case syscall.SIGWINCH:
			if rows, cols, err := termSize(); err == nil {
				_ = p.Resize(rows, cols)
			}
		case syscall.SIGINT:
			_, _ = p.Write([]byte{0x03})
		}
	}
}

func isReadDone(err error) bool {
	if errors.Is(err, io.EOF) {
		return true
	}
	var errno syscall.Errno
	if errors.As(err, &errno) && errno == syscall.EIO {
		return true
	}
	return false
}
