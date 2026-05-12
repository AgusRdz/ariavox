// Package tts provides a common interface for OS text-to-speech engines.
// Platform-specific implementations use build tags.
package tts

import "github.com/AgusRdz/ariavox/internal/processor"

// Speaker is the common TTS interface implemented per-platform.
type Speaker interface {
	Speak(text string, priority processor.Priority) error
	Stop() error
	SetRate(rate int) error   // 0-100
	SetVolume(vol int) error  // 0-100
	Close() error
}
