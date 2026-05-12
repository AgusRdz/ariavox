//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
)

// configFilePath returns $XDG_CONFIG_HOME/ariavox/ariavox.yaml,
// defaulting to ~/.config/ariavox/ariavox.yaml.
func configFilePath() (string, error) {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		xdg = filepath.Join(home, ".config")
	}
	return filepath.Join(xdg, "ariavox", "ariavox.yaml"), nil
}
