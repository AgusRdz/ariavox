package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/AgusRdz/ariavox/internal/config"
)

type depCheck struct {
	name     string
	commands []string // try in order
	required bool
	platform string // "" = all, "darwin", "linux", "windows"
}

var deps = []depCheck{
	{name: "say (TTS macOS)", commands: []string{"say"}, required: false, platform: "darwin"},
	{name: "espeak-ng (TTS Linux)", commands: []string{"espeak-ng", "espeak"}, required: false, platform: "linux"},
	{name: "spd-say (speech-dispatcher)", commands: []string{"spd-say"}, required: false, platform: "linux"},
	{name: "powershell (TTS Windows)", commands: []string{"powershell"}, required: false, platform: "windows"},
	{name: "curl", commands: []string{"curl"}, required: false, platform: ""},
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Verify system dependencies for ariavox",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("ariavox doctor — platform: %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

			allOK := true

			// dependency checks
			fmt.Println("dependencies:")
			for _, dep := range deps {
				if dep.platform != "" && dep.platform != runtime.GOOS {
					continue
				}
				found := ""
				for _, c := range dep.commands {
					if p, err := exec.LookPath(c); err == nil {
						found = p
						break
					}
				}
				if found != "" {
					fmt.Printf("  ok      %s (%s)\n", dep.name, found)
				} else if dep.required {
					fmt.Printf("  MISSING %s (required)\n", dep.name)
					allOK = false
				} else {
					fmt.Printf("  --      %s (optional, not found)\n", dep.name)
				}
			}

			// config file check
			fmt.Println("\nconfig:")
			cfgPath, err := configFilePath()
			if err != nil {
				fmt.Printf("  error   could not determine config path: %v\n", err)
				allOK = false
			} else {
				if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
					fmt.Printf("  --      %s (not found, defaults will be used)\n", cfgPath)
				} else if err != nil {
					fmt.Printf("  error   %s: %v\n", cfgPath, err)
					allOK = false
				} else {
					if _, err := config.Load(cfgPath); err != nil {
						fmt.Printf("  invalid %s: %v\n", cfgPath, err)
						allOK = false
					} else {
						fmt.Printf("  ok      %s\n", cfgPath)
					}
				}
			}

			// TTS voices
			fmt.Println("\nTTS voices:")
			printVoices()

			// environment
			fmt.Println("\nenvironment:")
			if v := os.Getenv("ARIAVOX_SR"); v != "" {
				fmt.Printf("  ARIAVOX_SR=%s (screen reader mode forced on)\n", v)
			} else {
				fmt.Println("  ARIAVOX_SR not set")
			}
			if v := os.Getenv("NO_COLOR"); v != "" {
				fmt.Printf("  NO_COLOR=%s (high contrast mode active)\n", v)
			} else {
				fmt.Println("  NO_COLOR not set")
			}

			fmt.Println()
			if allOK {
				fmt.Println("all checks passed")
			} else {
				fmt.Println("some checks failed — see above")
				return fmt.Errorf("doctor: checks failed")
			}
			return nil
		},
	}
}
