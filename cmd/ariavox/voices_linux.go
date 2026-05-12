//go:build linux

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func printVoices() {
	// Try spd-say first
	if _, err := exec.LookPath("spd-say"); err == nil {
		out, err := exec.Command("spd-say", "--list-synthesis-voices").Output()
		if err == nil && len(out) > 0 {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			fmt.Printf("  %d voices via speech-dispatcher. To set one:\n", len(lines))
			fmt.Println("    ariavox config set tts.voice \"<name>\"")
			fmt.Println()
			for _, l := range lines {
				fmt.Println("  " + l)
			}
			return
		}
	}

	// Try espeak-ng
	if _, err := exec.LookPath("espeak-ng"); err == nil {
		out, err := exec.Command("espeak-ng", "--voices").Output()
		if err == nil && len(out) > 0 {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			fmt.Printf("  %d voices via espeak-ng. To set one:\n", len(lines))
			fmt.Println("    ariavox config set tts.voice \"<language-code>\"")
			fmt.Println("    example: ariavox config set tts.voice \"es\"")
			fmt.Println()
			// Show first 20 lines only
			for i, l := range lines {
				if i >= 20 {
					fmt.Printf("  ... and %d more (run: espeak-ng --voices)\n", len(lines)-20)
					break
				}
				fmt.Println("  " + l)
			}
			return
		}
	}

	fmt.Println("  no TTS backend found.")
	fmt.Println()
	fmt.Println("  install one of:")
	fmt.Println("    sudo apt install speech-dispatcher   # spd-say (recommended)")
	fmt.Println("    sudo apt install espeak-ng           # espeak-ng (fallback)")
}
