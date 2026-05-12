// Package config loads and validates the ariavox YAML configuration.
package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/AgusRdz/ariavox/internal/processor"
)

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
	Enabled          bool `yaml:"enabled"`
	StripANSI        bool `yaml:"strip_ansi"`
	SuppressSpinners bool `yaml:"suppress_spinners"`
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
	TTS          TTSConfig          `yaml:"tts"`
	ScreenReader ScreenReaderConfig `yaml:"screen_reader"`
	Renderer     RendererConfig     `yaml:"renderer"`
	Agent        AgentConfig        `yaml:"agent"`
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

// Load reads a YAML config file and merges it over the defaults.
// Returns Default() if path does not exist.
func Load(path string) (Config, error) {
	cfg := Default()

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("config: open %s: %w", path, err)
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("config: parse %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

// Validate checks that all values are within acceptable ranges.
func (c *Config) Validate() error {
	if c.TTS.Rate < 0 || c.TTS.Rate > 100 {
		return fmt.Errorf("tts.rate must be 0-100, got %d", c.TTS.Rate)
	}
	if c.TTS.Volume < 0 || c.TTS.Volume > 100 {
		return fmt.Errorf("tts.volume must be 0-100, got %d", c.TTS.Volume)
	}
	for _, k := range c.TTS.PriorityFilter {
		if _, ok := kindFromString(k); !ok {
			return fmt.Errorf("tts.priority_filter: unknown kind %q", k)
		}
	}
	return nil
}

// TTSKindFilter converts PriorityFilter strings to EventKind values.
// Returns nil when the filter is empty (meaning: speak everything).
func (c *Config) TTSKindFilter() []processor.EventKind {
	if len(c.TTS.PriorityFilter) == 0 {
		return nil
	}
	kinds := make([]processor.EventKind, 0, len(c.TTS.PriorityFilter))
	for _, s := range c.TTS.PriorityFilter {
		if k, ok := kindFromString(s); ok {
			kinds = append(kinds, k)
		}
	}
	return kinds
}

// Set updates a single dotted key (e.g. "tts.rate") to the given string value
// in the YAML file at path, preserving all other content and comments.
func Set(path, key, value string) error {
	// read existing file or start empty
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("config: read %s: %w", path, err)
	}

	var root yaml.Node
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("config: parse %s: %w", path, err)
		}
	}
	if root.Kind == 0 {
		root = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{
			{Kind: yaml.MappingNode},
		}}
	}

	parts := strings.SplitN(key, ".", 2)
	if err := setNode(root.Content[0], parts, value); err != nil {
		return fmt.Errorf("config: set %s: %w", key, err)
	}

	out, err := yaml.Marshal(root.Content[0])
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	return os.WriteFile(path, out, 0600)
}

// setNode recursively walks/creates YAML mapping nodes to set a value.
func setNode(node *yaml.Node, parts []string, value string) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node")
	}
	key := parts[0]
	// find existing key
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			if len(parts) == 1 {
				node.Content[i+1].Value = value
				node.Content[i+1].Tag = inferTag(value)
				return nil
			}
			return setNode(node.Content[i+1], parts[1:], value)
		}
	}
	// key not found — create it
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	if len(parts) == 1 {
		valNode := &yaml.Node{Kind: yaml.ScalarNode, Value: value, Tag: inferTag(value)}
		node.Content = append(node.Content, keyNode, valNode)
		return nil
	}
	child := &yaml.Node{Kind: yaml.MappingNode}
	node.Content = append(node.Content, keyNode, child)
	return setNode(child, parts[1:], value)
}

// inferTag guesses the YAML scalar tag from the string value.
func inferTag(v string) string {
	switch strings.ToLower(v) {
	case "true", "false":
		return "!!bool"
	}
	// check numeric
	allDigits := true
	hasDot := false
	for i, c := range v {
		if c == '-' && i == 0 {
			continue
		}
		if c == '.' && !hasDot {
			hasDot = true
			continue
		}
		if c < '0' || c > '9' {
			allDigits = false
			break
		}
	}
	if allDigits && v != "" && v != "-" {
		if hasDot {
			return "!!float"
		}
		return "!!int"
	}
	return "!!str"
}

// kindFromString maps a priority_filter string to its EventKind.
var kindNames = map[string]processor.EventKind{
	"text":        processor.EventText,
	"code":        processor.EventCode,
	"tool_use":    processor.EventToolUse,
	"tool_result": processor.EventToolResult,
	"error":       processor.EventError,
	"task_start":  processor.EventTaskStart,
	"task_end":    processor.EventTaskEnd,
	"spinner":     processor.EventSpinner,
	"clear":       processor.EventClearScreen,
}

func kindFromString(s string) (processor.EventKind, bool) {
	k, ok := kindNames[s]
	return k, ok
}

// KindNames returns the valid kind name strings for use in error messages.
func KindNames() []string {
	names := make([]string, 0, len(kindNames))
	for k := range kindNames {
		names = append(names, k)
	}
	return names
}
