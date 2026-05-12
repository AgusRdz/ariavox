package ansi_test

import (
	"testing"

	"github.com/AgusRdz/ariavox/pkg/ansi"
)

func TestStripColors(t *testing.T) {
	input := []byte("\x1b[31mhello\x1b[0m world")
	got := string(ansi.Strip(input))
	want := "hello world"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestStripCursorMove(t *testing.T) {
	input := []byte("foo\x1b[Abar")
	got := string(ansi.Strip(input))
	want := "foobar"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestFragmentedSequence(t *testing.T) {
	p := &ansi.Parser{}
	// Send ESC[31m in two fragments
	out := p.Process([]byte("\x1b[3"))
	out = append(out, p.Process([]byte("1mtext"))...)
	if string(out) != "text" {
		t.Errorf("got %q want %q", string(out), "text")
	}
}

func TestPlainText(t *testing.T) {
	input := []byte("hello world")
	got := string(ansi.Strip(input))
	if got != "hello world" {
		t.Errorf("got %q want %q", got, "hello world")
	}
}
