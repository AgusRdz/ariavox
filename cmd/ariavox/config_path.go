package main

import (
	"os"
	"path/filepath"
	"runtime"
)

// configFilePath returns the platform-appropriate config file path.
//
// macOS/Linux: $XDG_CONFIG_HOME/ariavox/ariavox.yaml (defaults to ~/.config)
// Windows:     %APPDATA%\ariavox\ariavox.yaml
func configFilePath() (string, error) {
	var dir string
	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			appdata = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		dir = filepath.Join(appdata, "ariavox")
	default:
		xdg := os.Getenv("XDG_CONFIG_HOME")
		if xdg == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			xdg = filepath.Join(home, ".config")
		}
		dir = filepath.Join(xdg, "ariavox")
	}
	return filepath.Join(dir, "ariavox.yaml"), nil
}

const defaultConfig = `# ariavox configuration
# Edit with: ariavox config edit
# Full reference: https://github.com/AgusRdz/ariavox

tts:
  enabled: false
  rate: 50        # 0-100
  volume: 80      # 0-100
  voice: ""       # empty = system default
  priority_filter:
    - task_end
    - error
    - tool_use

screen_reader:
  enabled: false        # set true or use --sr flag / ARIAVOX_SR=1
  strip_ansi: true
  suppress_spinners: true

renderer:
  separators: true
  semantic_prefixes: true
  high_contrast: false
  respect_no_color: true

agent:
  command: ["claude"]
  args: []
`

func ensureDefaultConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(defaultConfig)
	return err
}
