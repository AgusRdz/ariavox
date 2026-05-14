//go:build windows

package main

import "os"

// replaceExe replaces dst with src on Windows.
// Windows does not allow overwriting a running executable directly, but it
// does allow renaming it. Strategy: rename current exe to .old, then rename
// new binary into place. cleanOldExe() removes the .old on next startup.
func replaceExe(src, dst string) error {
	old := dst + ".old"
	_ = os.Remove(old) // remove any stale backup from a previous update
	if err := os.Rename(dst, old); err != nil {
		return err
	}
	if err := os.Rename(src, dst); err != nil {
		// try to restore the original
		_ = os.Rename(old, dst)
		return err
	}
	return nil
}
