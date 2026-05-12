package main

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
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
					fmt.Printf("  ok  %s (%s)\n", dep.name, found)
				} else if dep.required {
					fmt.Printf("  MISSING (required)  %s\n", dep.name)
					allOK = false
				} else {
					fmt.Printf("  --  %s (optional, not found)\n", dep.name)
				}
			}

			fmt.Println()
			if allOK {
				fmt.Println("all required dependencies satisfied")
			} else {
				fmt.Println("some required dependencies are missing")
			}
			return nil
		},
	}
}
