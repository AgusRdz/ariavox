//go:build windows

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
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
}

func handleSignals(ch <-chan os.Signal, p pty.PTY) {
	for sig := range ch {
		if sig == syscall.SIGINT {
			_, _ = p.Write([]byte{0x03})
		}
	}
}

func isReadDone(err error) bool {
	return errors.Is(err, io.EOF)
}
