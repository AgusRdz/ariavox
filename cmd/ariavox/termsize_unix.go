//go:build darwin || linux

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func termSize() (rows, cols uint16, err error) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 24, 80, err
	}
	return ws.Row, ws.Col, nil
}
