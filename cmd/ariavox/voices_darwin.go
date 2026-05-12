//go:build darwin

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func printVoices() {
	out, err := exec.Command("say", "-v", "?").Output()
	if err != nil {
		fmt.Println("  (could not query voices — is 'say' available?)")
		fmt.Println()
		fmt.Println("  to set a voice:")
		fmt.Println("    ariavox config set tts.voice \"Paulina\"")
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		fmt.Println("  (no voices found)")
		return
	}

	fmt.Printf("  %d voices installed. To set one:\n", len(lines))
	fmt.Println("    ariavox config set tts.voice \"<name>\"")
	fmt.Println()
	fmt.Println("  tip: download Enhanced/Premium voices in")
	fmt.Println("  System Settings → Accessibility → Vision → Read & Speak → System Voice → (i)")
	fmt.Println()

	// Show voices grouped: first non-English, then English
	var spanish, other, english []string
	for _, line := range lines {
		name := strings.Fields(line)[0]
		lower := strings.ToLower(line)
		if strings.Contains(lower, "spanish") || strings.Contains(lower, "es_") {
			spanish = append(spanish, "  "+line)
		} else if strings.Contains(lower, "english") || strings.Contains(lower, "en_") {
			english = append(english, "  "+line)
		} else {
			other = append(other, "  "+line)
		}
		_ = name
	}

	if len(spanish) > 0 {
		fmt.Println("  spanish voices:")
		for _, v := range spanish {
			fmt.Println(v)
		}
		fmt.Println()
	}
	if len(english) > 0 {
		fmt.Printf("  english voices (%d):\n", len(english))
		// only show first 5 to keep output manageable
		for i, v := range english {
			if i >= 5 {
				fmt.Printf("    ... and %d more (run: say -v '?' | grep -i english)\n", len(english)-5)
				break
			}
			fmt.Println(v)
		}
		fmt.Println()
	}
	if len(other) > 0 {
		fmt.Printf("  other languages (%d): run 'say -v ?' to list all\n", len(other))
	}
}
