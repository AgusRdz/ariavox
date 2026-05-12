package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AgusRdz/ariavox/internal/config"
	"github.com/AgusRdz/ariavox/internal/processor"
)

func writeTmp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "ariavox-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.TTS.Rate != 25 {
		t.Errorf("default TTS.Rate = %d, want 25", cfg.TTS.Rate)
	}
	if cfg.TTS.Volume != 80 {
		t.Errorf("default TTS.Volume = %d, want 80", cfg.TTS.Volume)
	}
	if !cfg.Renderer.Separators {
		t.Error("default Renderer.Separators should be true")
	}
}

func TestLoad_NonExistentReturnsDefault(t *testing.T) {
	cfg, err := config.Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	def := config.Default()
	if cfg.TTS.Rate != def.TTS.Rate {
		t.Errorf("got TTS.Rate=%d, want %d", cfg.TTS.Rate, def.TTS.Rate)
	}
}

func TestLoad_ParsesValues(t *testing.T) {
	path := writeTmp(t, `
tts:
  enabled: true
  rate: 75
  volume: 60
renderer:
  high_contrast: true
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.TTS.Enabled {
		t.Error("TTS.Enabled should be true")
	}
	if cfg.TTS.Rate != 75 {
		t.Errorf("TTS.Rate = %d, want 75", cfg.TTS.Rate)
	}
	if cfg.TTS.Volume != 60 {
		t.Errorf("TTS.Volume = %d, want 60", cfg.TTS.Volume)
	}
	if !cfg.Renderer.HighContrast {
		t.Error("Renderer.HighContrast should be true")
	}
}

func TestLoad_InvalidRateReturnsError(t *testing.T) {
	path := writeTmp(t, "tts:\n  rate: 200\n")
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for rate=200")
	}
}

func TestLoad_InvalidKindReturnsError(t *testing.T) {
	path := writeTmp(t, "tts:\n  priority_filter:\n    - unknown_kind\n")
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for unknown kind in priority_filter")
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := config.Default()
	if err := cfg.Validate(); err != nil {
		t.Errorf("Default config should be valid: %v", err)
	}
}

func TestTTSKindFilter(t *testing.T) {
	cfg := config.Default()
	cfg.TTS.PriorityFilter = []string{"error", "tool_use"}

	kinds := cfg.TTSKindFilter()
	if len(kinds) != 2 {
		t.Fatalf("expected 2 kinds, got %d", len(kinds))
	}
	has := func(k processor.EventKind) bool {
		for _, v := range kinds {
			if v == k {
				return true
			}
		}
		return false
	}
	if !has(processor.EventError) {
		t.Error("expected EventError in filter")
	}
	if !has(processor.EventToolUse) {
		t.Error("expected EventToolUse in filter")
	}
}

func TestTTSKindFilter_EmptyReturnsNil(t *testing.T) {
	cfg := config.Default()
	cfg.TTS.PriorityFilter = nil
	if cfg.TTSKindFilter() != nil {
		t.Error("empty filter should return nil (speak everything)")
	}
}

func TestSet_CreatesKey(t *testing.T) {
	path := writeTmp(t, "tts:\n  rate: 50\n")
	if err := config.Set(path, "tts.rate", "70"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load after Set: %v", err)
	}
	if cfg.TTS.Rate != 70 {
		t.Errorf("after Set tts.rate=70, got %d", cfg.TTS.Rate)
	}
}

func TestSet_CreatesMissingNestedKey(t *testing.T) {
	path := writeTmp(t, "")
	if err := config.Set(path, "tts.volume", "55"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load after Set: %v", err)
	}
	if cfg.TTS.Volume != 55 {
		t.Errorf("after Set tts.volume=55, got %d", cfg.TTS.Volume)
	}
}

func TestSet_BoolValue(t *testing.T) {
	path := writeTmp(t, "tts:\n  enabled: false\n")
	if err := config.Set(path, "tts.enabled", "true"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load after Set: %v", err)
	}
	if !cfg.TTS.Enabled {
		t.Error("after Set tts.enabled=true, got false")
	}
}
