//go:build windows

package tts

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/AgusRdz/ariavox/internal/processor"
)

// WindowsSpeaker uses PowerShell + SAPI5.
type WindowsSpeaker struct {
	mu      sync.Mutex
	current *exec.Cmd
	rate    int
	volume  int
}

// New returns a Speaker for Windows.
func New() Speaker {
	return &WindowsSpeaker{rate: 50, volume: 80}
}

func (s *WindowsSpeaker) Speak(text string, priority processor.Priority) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if priority == processor.PriorityHigh && s.current != nil {
		s.current.Process.Kill() //nolint:errcheck
		s.current = nil
	}

	// Escape single quotes to avoid PowerShell injection
	safe := strings.ReplaceAll(text, "'", "''")
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Speech
$s = New-Object System.Speech.Synthesis.SpeechSynthesizer
$s.Rate = %d
$s.Volume = %d
$s.Speak('%s')
`, s.sapiRate(), s.volume, safe)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	s.current = cmd
	return cmd.Start()
}

func (s *WindowsSpeaker) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != nil {
		err := s.current.Process.Kill()
		s.current = nil
		return err
	}
	return nil
}

func (s *WindowsSpeaker) SetRate(rate int) error {
	if rate < 0 || rate > 100 {
		return fmt.Errorf("tts: rate must be 0-100")
	}
	s.rate = rate
	return nil
}

func (s *WindowsSpeaker) SetVolume(vol int) error {
	if vol < 0 || vol > 100 {
		return fmt.Errorf("tts: volume must be 0-100")
	}
	s.volume = vol
	return nil
}

func (s *WindowsSpeaker) Close() error { return s.Stop() }

// sapiRate converts 0-100 to SAPI5 rate (-10 to +10).
func (s *WindowsSpeaker) sapiRate() int {
	return -10 + (s.rate * 20 / 100)
}
