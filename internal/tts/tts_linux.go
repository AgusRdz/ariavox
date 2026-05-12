//go:build linux

package tts

import (
	"fmt"
	"os/exec"
	"strconv"
	"sync"

	"github.com/AgusRdz/ariavox/internal/processor"
)

// LinuxSpeaker tries speech-dispatcher (spd-say) first, falls back to espeak-ng.
type LinuxSpeaker struct {
	mu      sync.Mutex
	current *exec.Cmd
	rate    int
	volume  int
	voice   string
	backend string // detected at first Speak call
}

// New returns a Speaker for Linux.
func New() Speaker {
	return &LinuxSpeaker{rate: 50, volume: 80}
}

func (s *LinuxSpeaker) Speak(text string, priority processor.Priority) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.backend == "" {
		s.backend = detectLinuxBackend()
	}

	if priority == processor.PriorityHigh && s.current != nil {
		s.current.Process.Kill() //nolint:errcheck
		s.current = nil
	}

	var cmd *exec.Cmd
	switch s.backend {
	case "spd-say":
		cmd = exec.Command("spd-say", "-r", strconv.Itoa(s.rate), "-w", text)
	case "espeak-ng":
		cmd = exec.Command("espeak-ng", "-s", strconv.Itoa(s.espeakRate()), text)
	case "espeak":
		cmd = exec.Command("espeak", "-s", strconv.Itoa(s.espeakRate()), text)
	default:
		return fmt.Errorf("tts: no TTS backend found (install espeak-ng or speech-dispatcher)")
	}

	s.current = cmd
	return cmd.Start()
}

func (s *LinuxSpeaker) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != nil {
		err := s.current.Process.Kill()
		s.current = nil
		return err
	}
	return nil
}

func (s *LinuxSpeaker) SetRate(rate int) error {
	if rate < 0 || rate > 100 {
		return fmt.Errorf("tts: rate must be 0-100")
	}
	s.rate = rate
	return nil
}

func (s *LinuxSpeaker) SetVolume(vol int) error {
	if vol < 0 || vol > 100 {
		return fmt.Errorf("tts: volume must be 0-100")
	}
	s.volume = vol
	return nil
}

func (s *LinuxSpeaker) SetVoice(voice string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.voice = voice
}

func (s *LinuxSpeaker) Wait() error {
	s.mu.Lock()
	cmd := s.current
	s.mu.Unlock()
	if cmd != nil {
		return cmd.Wait()
	}
	return nil
}

func (s *LinuxSpeaker) Close() error { return s.Stop() }

func (s *LinuxSpeaker) espeakRate() int {
	return 80 + (s.rate * 320 / 100)
}

// detectLinuxBackend finds the first available TTS backend at runtime.
func detectLinuxBackend() string {
	for _, b := range []string{"spd-say", "espeak-ng", "espeak"} {
		if _, err := exec.LookPath(b); err == nil {
			return b
		}
	}
	return ""
}
