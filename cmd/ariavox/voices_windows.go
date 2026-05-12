//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func printVoices() {
	script := `Add-Type -AssemblyName System.Speech
$s = New-Object System.Speech.Synthesis.SpeechSynthesizer
$s.GetInstalledVoices() | ForEach-Object { $_.VoiceInfo.Name }`

	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil || len(out) == 0 {
		fmt.Println("  (could not query SAPI5 voices)")
		fmt.Println()
		fmt.Println("  to add voices: Settings → Time & Language → Speech → Add voices")
		fmt.Println()
		fmt.Println("  to set a voice:")
		fmt.Println("    ariavox config set tts.voice \"Microsoft Sabina Desktop\"")
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	fmt.Printf("  %d SAPI5 voices installed. To set one:\n", len(lines))
	fmt.Println("    ariavox config set tts.voice \"<name>\"")
	fmt.Println()
	for _, l := range lines {
		fmt.Println("  " + strings.TrimSpace(l))
	}
	fmt.Println()
	fmt.Println("  tip: add more voices in Settings → Time & Language → Speech → Add voices")
}
