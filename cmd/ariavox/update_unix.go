//go:build !windows

package main

import "os"

// replaceExe atomically replaces dst with src via rename (safe on Unix).
func replaceExe(src, dst string) error {
	return os.Rename(src, dst)
}
