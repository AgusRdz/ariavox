//go:build windows

package main

import (
	"os"
	"path/filepath"
)

// configFilePath returns %APPDATA%\ariavox\ariavox.yaml.
func configFilePath() (string, error) {
	appdata := os.Getenv("APPDATA")
	if appdata == "" {
		appdata = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}
	return filepath.Join(appdata, "ariavox", "ariavox.yaml"), nil
}
