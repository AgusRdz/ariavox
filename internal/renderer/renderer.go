// Package renderer writes accessible output to the terminal.
package renderer

import (
	"io"
	"os"

	"github.com/AgusRdz/ariavox/internal/processor"
	"github.com/AgusRdz/ariavox/pkg/ansi"
)

// Config controls renderer behavior.
type Config struct {
	SRMode           bool // strip all ANSI, suppress spinners
	Separators       bool
	SemanticPrefixes bool
	HighContrast     bool
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
	cfg Config
	out io.Writer
}

// New creates a Renderer writing to out.
func New(cfg Config, out io.Writer) *Renderer {
	if out == nil {
		out = os.Stdout
	}
	return &Renderer{cfg: cfg, out: out}
}

// Render writes the event to the output according to the current config.
func (r *Renderer) Render(e processor.Event) {
	text := e.Text
	if r.cfg.SRMode {
		text = string(ansi.Strip([]byte(text)))
	}
	if text == "" {
		return
	}
	r.out.Write([]byte(text)) //nolint:errcheck
}
