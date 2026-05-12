// Package config loads and validates the ariavox YAML configuration.
// Phase 5 stub — parsing wired when config commands are implemented.
package config

// TTSConfig holds TTS preferences.
type TTSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Rate           int      `yaml:"rate"`
	Volume         int      `yaml:"volume"`
	Voice          string   `yaml:"voice"`
	PriorityFilter []string `yaml:"priority_filter"`
}

// ScreenReaderConfig holds screen reader mode preferences.
type ScreenReaderConfig struct {
	Enabled           bool `yaml:"enabled"`
	StripANSI         bool `yaml:"strip_ansi"`
	SuppressSpinners  bool `yaml:"suppress_spinners"`
}

// RendererConfig holds rendering preferences.
type RendererConfig struct {
	Separators       bool `yaml:"separators"`
	SemanticPrefixes bool `yaml:"semantic_prefixes"`
	HighContrast     bool `yaml:"high_contrast"`
	RespectNoColor   bool `yaml:"respect_no_color"`
}

// AgentConfig holds the wrapped agent command.
type AgentConfig struct {
	Command []string `yaml:"command"`
	Args    []string `yaml:"args"`
}

// Config is the root configuration structure.
type Config struct {
	TTS           TTSConfig          `yaml:"tts"`
	ScreenReader  ScreenReaderConfig `yaml:"screen_reader"`
	Renderer      RendererConfig     `yaml:"renderer"`
	Agent         AgentConfig        `yaml:"agent"`
}

// Default returns a Config with safe defaults.
func Default() Config {
	return Config{
		TTS: TTSConfig{
			Enabled: false,
			Rate:    50,
			Volume:  80,
			PriorityFilter: []string{"task_end", "error", "tool_use"},
		},
		ScreenReader: ScreenReaderConfig{
			StripANSI:        true,
			SuppressSpinners: true,
		},
		Renderer: RendererConfig{
			Separators:       true,
			SemanticPrefixes: true,
			RespectNoColor:   true,
		},
		Agent: AgentConfig{
			Command: []string{"claude"},
		},
	}
}
