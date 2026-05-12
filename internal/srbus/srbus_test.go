package srbus_test

import (
	"testing"

	"github.com/AgusRdz/ariavox/internal/srbus"
)

func TestNopBus_AnnounceAlwaysNil(t *testing.T) {
	b := srbus.NopBus{}
	if err := b.Announce("hello"); err != nil {
		t.Errorf("NopBus.Announce: %v", err)
	}
	if err := b.Announce(""); err != nil {
		t.Errorf("NopBus.Announce empty: %v", err)
	}
}

func TestNopBus_CloseAlwaysNil(t *testing.T) {
	b := srbus.NopBus{}
	if err := b.Close(); err != nil {
		t.Errorf("NopBus.Close: %v", err)
	}
}

// TestNew_ReturnsBus verifies New() always returns a non-nil Bus regardless
// of whether the platform SR infrastructure is available.
func TestNew_ReturnsBus(t *testing.T) {
	b := srbus.New()
	if b == nil {
		t.Fatal("srbus.New() returned nil")
	}
	defer b.Close()

	// Announce must not panic, even if the underlying bus is unavailable.
	// Errors are acceptable (degrade silently).
	_ = b.Announce("ariavox test announcement")
}

// TestNew_EmptyTextSafe verifies empty text does not panic or error.
func TestNew_EmptyTextSafe(t *testing.T) {
	b := srbus.New()
	defer b.Close()
	if err := b.Announce(""); err != nil {
		t.Errorf("Announce empty string returned error: %v", err)
	}
}

// TestNew_CloseIdempotent verifies Close does not panic on repeated calls.
func TestNew_CloseIdempotent(t *testing.T) {
	b := srbus.New()
	if err := b.Close(); err != nil {
		t.Logf("first Close: %v (acceptable if bus not available)", err)
	}
	// second close must not panic
	_ = b.Close()
}
