// Package renderer writes accessible output to the terminal.
package renderer

import (
	"io"
	"os"
	"strings"

	"github.com/AgusRdz/ariavox/internal/processor"
	"github.com/AgusRdz/ariavox/pkg/ansi"
)

// Config controls renderer behavior.
type Config struct {
	SRMode           bool // strip all ANSI, suppress spinners
	Separators       bool // blank line between different event types
	SemanticPrefixes bool // prepend kind label before each event
	HighContrast     bool // replace Unicode decoration with ASCII equivalents
}

// DefaultConfig returns safe defaults.
func DefaultConfig() Config {
	return Config{
		Separators:       true,
		SemanticPrefixes: true,
	}
}

// Renderer writes events to an output writer in accessible form.
type Renderer struct {
	cfg      Config
	out      io.Writer
	lastKind processor.EventKind
	firstEv  bool // tracks whether any event has been written yet
}

// New creates a Renderer writing to out.
func New(cfg Config, out io.Writer) *Renderer {
	if out == nil {
		out = os.Stdout
	}
	return &Renderer{cfg: cfg, out: out, firstEv: true}
}

// Render writes the event to the output according to the current config.
func (r *Renderer) Render(e processor.Event) {
	// suppress spinners in SR mode
	if r.cfg.SRMode && e.Kind == processor.EventSpinner {
		return
	}

	text := e.Text
	if r.cfg.SRMode || r.cfg.HighContrast {
		text = string(ansi.Strip([]byte(text)))
	}
	if r.cfg.HighContrast {
		text = toASCII(text)
	}

	// TaskStart/TaskEnd emit no visible text of their own
	if e.Kind == processor.EventTaskStart {
		if r.cfg.Separators && !r.firstEv {
			r.write("\n")
		}
		r.firstEv = false
		r.lastKind = e.Kind
		return
	}
	if e.Kind == processor.EventTaskEnd {
		if r.cfg.Separators {
			r.write("\n")
		}
		r.lastKind = e.Kind
		return
	}

	if text == "" {
		return
	}

	// separator between transitions to a different event kind
	if r.cfg.Separators && !r.firstEv && e.Kind != r.lastKind {
		r.write("\n")
	}

	// semantic prefix
	if r.cfg.SemanticPrefixes {
		if prefix := kindPrefix(e.Kind, r.cfg.HighContrast); prefix != "" {
			r.write(prefix)
		}
	}

	r.write(text)
	r.write("\n")

	r.firstEv = false
	r.lastKind = e.Kind
}

// Reset clears state between sessions (e.g. after child process exits).
func (r *Renderer) Reset() {
	r.firstEv = true
	r.lastKind = 0
}

func (r *Renderer) write(s string) {
	_, _ = io.WriteString(r.out, s)
}

// kindPrefix returns the accessible text label for an event kind.
// Returns empty string for kinds that need no prefix.
func kindPrefix(kind processor.EventKind, highContrast bool) string {
	switch kind {
	case processor.EventCode:
		if highContrast {
			return "[code] "
		}
		return "⟨code⟩ "
	case processor.EventToolUse:
		if highContrast {
			return "[tool] "
		}
		return "⟨tool⟩ "
	case processor.EventToolResult:
		if highContrast {
			return "[result] "
		}
		return "⟨result⟩ "
	case processor.EventError:
		if highContrast {
			return "[error] "
		}
		return "⟨error⟩ "
	case processor.EventSpinner:
		if highContrast {
			return "[thinking] "
		}
		return "⟨thinking⟩ "
	default:
		return ""
	}
}

// toASCII replaces common Unicode decoration characters with ASCII equivalents.
func toASCII(s string) string {
	replacer := strings.NewReplacer(
		// Braille spinner chars → empty (suppressed in high-contrast)
		"⠋", "", "⠙", "", "⠹", "", "⠸", "", "⠼", "",
		"⠴", "", "⠦", "", "⠧", "", "⠇", "", "⠏", "",
		// Claude Code tool decorators
		"●", "*",
		"⎿", "->",
		// Smart quotes and dashes
		"‘", "'", "’", "'",
		"“", "\"", "”", "\"",
		"–", "-", "—", "--",
		// Code fence brackets
		"⟨", "<", "⟩", ">",
	)
	return replacer.Replace(s)
}
