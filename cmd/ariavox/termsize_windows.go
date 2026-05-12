//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func termSize() (rows, cols uint16, err error) {
	var info windows.ConsoleScreenBufferInfo
	h := windows.Handle(os.Stdout.Fd())
	if err := windows.GetConsoleScreenBufferInfo(h, &info); err != nil {
		return 24, 80, err
	}
	cols = uint16(info.Window.Right - info.Window.Left + 1)
	rows = uint16(info.Window.Bottom - info.Window.Top + 1)
	return rows, cols, nil
}
