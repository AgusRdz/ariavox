//go:build darwin

package tts

import (
	"fmt"
	"os/exec"
	"strconv"
	"sync"

	"github.com/AgusRdz/ariavox/internal/processor"
)

// DarwinSpeaker uses the macOS `say` command.
type DarwinSpeaker struct {
	mu      sync.Mutex
	current *exec.Cmd
	rate    int
	volume  int
	voice   string
}

// New returns a Speaker for macOS.
func New() Speaker {
	return &DarwinSpeaker{rate: 50, volume: 80}
}

func (s *DarwinSpeaker) Speak(text string, priority processor.Priority) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if priority == processor.PriorityHigh && s.current != nil {
		s.current.Process.Kill() //nolint:errcheck
		s.current = nil
	}

	args := []string{"-r", strconv.Itoa(s.rateToWPM())}
	if s.voice != "" {
		args = append(args, "-v", s.voice)
	}
	args = append(args, text)

	cmd := exec.Command("say", args...)
	s.current = cmd
	return cmd.Start()
}

func (s *DarwinSpeaker) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != nil {
		err := s.current.Process.Kill()
		s.current = nil
		return err
	}
	return nil
}

func (s *DarwinSpeaker) SetRate(rate int) error {
	if rate < 0 || rate > 100 {
		return fmt.Errorf("tts: rate must be 0-100")
	}
	s.rate = rate
	return nil
}

func (s *DarwinSpeaker) SetVolume(vol int) error {
	if vol < 0 || vol > 100 {
		return fmt.Errorf("tts: volume must be 0-100")
	}
	s.volume = vol
	return nil
}

func (s *DarwinSpeaker) SetVoice(voice string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.voice = voice
}

func (s *DarwinSpeaker) Wait() error {
	s.mu.Lock()
	cmd := s.current
	s.mu.Unlock()
	if cmd != nil {
		return cmd.Wait()
	}
	return nil
}

func (s *DarwinSpeaker) Close() error { return s.Stop() }

// rateToWPM converts 0-100 scale to words-per-minute for `say -r`.
// 0=120 wpm (slow), 25=160 wpm (default, natural), 50=200 wpm, 100=300 wpm.
func (s *DarwinSpeaker) rateToWPM() int {
	return 120 + (s.rate * 180 / 100)
}
