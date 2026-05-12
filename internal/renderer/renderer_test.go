package renderer_test

import (
	"strings"
	"testing"

	"github.com/AgusRdz/ariavox/internal/processor"
	"github.com/AgusRdz/ariavox/internal/renderer"
)

func render(cfg renderer.Config, events ...processor.Event) string {
	var buf strings.Builder
	r := renderer.New(cfg, &buf)
	for _, e := range events {
		r.Render(e)
	}
	return buf.String()
}

func ev(kind processor.EventKind, text string) processor.Event {
	return processor.Event{Kind: kind, Text: text, Priority: processor.PriorityMedium}
}

// --- SR mode ---

func TestSRMode_SuppressesSpinner(t *testing.T) {
	cfg := renderer.Config{SRMode: true}
	out := render(cfg, ev(processor.EventSpinner, "⠋ Thinking…"))
	if out != "" {
		t.Errorf("SRMode should suppress spinners, got %q", out)
	}
}

func TestSRMode_StripsANSI(t *testing.T) {
	cfg := renderer.Config{SRMode: true}
	out := render(cfg, ev(processor.EventText, "\x1b[32mhello\x1b[0m"))
	if strings.Contains(out, "\x1b") {
		t.Errorf("SRMode should strip ANSI, got %q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("SRMode should preserve text content, got %q", out)
	}
}

func TestSRMode_SpinnerNotSuppressedWhenDisabled(t *testing.T) {
	cfg := renderer.Config{SRMode: false}
	out := render(cfg, ev(processor.EventSpinner, "⠋ Thinking…"))
	if out == "" {
		t.Error("spinners should pass through when SRMode is off")
	}
}

// --- semantic prefixes ---

func TestSemanticPrefixes_Text(t *testing.T) {
	cfg := renderer.Config{SemanticPrefixes: true}
	out := render(cfg, ev(processor.EventText, "hello"))
	// text events have no prefix
	if strings.HasPrefix(out, "⟨") {
		t.Errorf("EventText should have no prefix, got %q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("text content missing, got %q", out)
	}
}

func TestSemanticPrefixes_Code(t *testing.T) {
	cfg := renderer.Config{SemanticPrefixes: true}
	out := render(cfg, ev(processor.EventCode, "```go"))
	if !strings.Contains(out, "⟨code⟩") {
		t.Errorf("EventCode should have ⟨code⟩ prefix, got %q", out)
	}
}

func TestSemanticPrefixes_ToolUse(t *testing.T) {
	cfg := renderer.Config{SemanticPrefixes: true}
	out := render(cfg, ev(processor.EventToolUse, "● Read(file.go)"))
	if !strings.Contains(out, "⟨tool⟩") {
		t.Errorf("EventToolUse should have ⟨tool⟩ prefix, got %q", out)
	}
}

func TestSemanticPrefixes_Error(t *testing.T) {
	cfg := renderer.Config{SemanticPrefixes: true}
	out := render(cfg, ev(processor.EventError, "Error: build failed"))
	if !strings.Contains(out, "⟨error⟩") {
		t.Errorf("EventError should have ⟨error⟩ prefix, got %q", out)
	}
}

func TestSemanticPrefixes_Disabled(t *testing.T) {
	cfg := renderer.Config{SemanticPrefixes: false}
	out := render(cfg, ev(processor.EventCode, "```go"))
	if strings.Contains(out, "⟨") {
		t.Errorf("prefixes should be absent when SemanticPrefixes=false, got %q", out)
	}
}

// --- separators ---

func TestSeparators_BetweenDifferentKinds(t *testing.T) {
	cfg := renderer.Config{Separators: true}
	out := render(cfg,
		ev(processor.EventText, "intro"),
		ev(processor.EventCode, "```go"),
	)
	// should have a blank line between text and code
	if !strings.Contains(out, "\n\n") {
		t.Errorf("expected blank line separator between different kinds, got %q", out)
	}
}

func TestSeparators_NotBetweenSameKind(t *testing.T) {
	cfg := renderer.Config{Separators: true}
	out := render(cfg,
		ev(processor.EventText, "line one"),
		ev(processor.EventText, "line two"),
	)
	// consecutive same-kind events should not get a double blank line
	if strings.Contains(out, "\n\n") {
		t.Errorf("consecutive same-kind events should not insert separator, got %q", out)
	}
}

func TestSeparators_Disabled(t *testing.T) {
	cfg := renderer.Config{Separators: false}
	out := render(cfg,
		ev(processor.EventText, "intro"),
		ev(processor.EventCode, "```go"),
	)
	if strings.Contains(out, "\n\n") {
		t.Errorf("separators should be absent when disabled, got %q", out)
	}
}

func TestTaskStart_InsertsBlankLine(t *testing.T) {
	cfg := renderer.Config{Separators: true}
	// two responses separated by TaskStart
	out := render(cfg,
		ev(processor.EventText, "first response"),
		processor.Event{Kind: processor.EventTaskStart},
		ev(processor.EventText, "second response"),
	)
	if !strings.Contains(out, "\n\n") {
		t.Errorf("TaskStart should insert blank line between responses, got %q", out)
	}
}

// --- high contrast ---

func TestHighContrast_ToolDecorators(t *testing.T) {
	cfg := renderer.Config{HighContrast: true, SemanticPrefixes: true}
	out := render(cfg, ev(processor.EventToolUse, "● Read(file.go)"))
	if strings.Contains(out, "●") {
		t.Errorf("high contrast should replace ● with *, got %q", out)
	}
	if !strings.Contains(out, "*") {
		t.Errorf("high contrast should have * in place of ●, got %q", out)
	}
	if !strings.Contains(out, "[tool]") {
		t.Errorf("high contrast prefix should be [tool], got %q", out)
	}
}

func TestHighContrast_CodePrefix(t *testing.T) {
	cfg := renderer.Config{HighContrast: true, SemanticPrefixes: true}
	out := render(cfg, ev(processor.EventCode, "```go"))
	if !strings.Contains(out, "[code]") {
		t.Errorf("high contrast code prefix should be [code], got %q", out)
	}
}

// --- empty text ---

func TestEmptyTextSkipped(t *testing.T) {
	cfg := renderer.DefaultConfig()
	out := render(cfg, ev(processor.EventText, ""))
	if out != "" {
		t.Errorf("empty text event should produce no output, got %q", out)
	}
}

// --- Reset ---

func TestReset_FirstEventHasNoLeadingSeparator(t *testing.T) {
	cfg := renderer.Config{Separators: true}
	var buf strings.Builder
	r := renderer.New(cfg, &buf)

	r.Render(ev(processor.EventText, "first"))
	r.Reset()
	buf.Reset()
	r.Render(ev(processor.EventText, "after reset"))

	out := buf.String()
	if strings.HasPrefix(out, "\n") {
		t.Errorf("first event after Reset should not have leading newline, got %q", out)
	}
}
