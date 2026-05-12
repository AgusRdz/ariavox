package tts_test

import (
	"sync"
	"testing"
	"time"

	"github.com/AgusRdz/ariavox/internal/processor"
	"github.com/AgusRdz/ariavox/internal/tts"
)

// mockSpeaker records calls and simulates instant speech.
type mockSpeaker struct {
	mu      sync.Mutex
	spoken  []string
	stopped int
}

func (m *mockSpeaker) Speak(text string, _ processor.Priority) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.spoken = append(m.spoken, text)
	return nil
}

func (m *mockSpeaker) Wait() error  { return nil }
func (m *mockSpeaker) Stop() error  { m.mu.Lock(); m.stopped++; m.mu.Unlock(); return nil }
func (m *mockSpeaker) SetRate(_ int) error   { return nil }
func (m *mockSpeaker) SetVolume(_ int) error { return nil }
func (m *mockSpeaker) Close() error { return m.Stop() }

func (m *mockSpeaker) getSpoken() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.spoken))
	copy(out, m.spoken)
	return out
}

func (m *mockSpeaker) getStopCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopped
}

func textEv(text string, priority processor.Priority) processor.Event {
	return processor.Event{Kind: processor.EventText, Text: text, Priority: priority}
}

func taskStart() processor.Event {
	return processor.Event{Kind: processor.EventTaskStart, Priority: processor.PriorityHigh}
}

func waitSpoken(t *testing.T, mock *mockSpeaker, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(mock.getSpoken()) >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d spoken items, got %d: %v", want, len(mock.getSpoken()), mock.getSpoken())
}

// --- tests ---

func TestBridge_MediumPriorityEnqueues(t *testing.T) {
	mock := &mockSpeaker{}
	b := tts.NewBridge(mock, nil)
	defer b.Close()

	b.Send(taskStart(), nil)
	b.Send(textEv("one", processor.PriorityMedium), nil)
	b.Send(textEv("two", processor.PriorityMedium), nil)
	b.Send(textEv("three", processor.PriorityMedium), nil)

	waitSpoken(t, mock, 3)

	spoken := mock.getSpoken()
	if len(spoken) != 3 {
		t.Fatalf("expected 3 items spoken, got %d: %v", len(spoken), spoken)
	}
}

func TestBridge_HighPriorityClearsQueue(t *testing.T) {
	// Use a blocking speaker so we can observe queue clearing.
	block := make(chan struct{})
	released := false

	blocking := &blockingSpeaker{block: block, mock: &mockSpeaker{}}
	b := tts.NewBridge(blocking, nil)
	defer b.Close()

	// send first item — consumer will block on Wait
	b.Send(taskStart(), nil)
	b.Send(textEv("queued-1", processor.PriorityMedium), nil)
	b.Send(textEv("queued-2", processor.PriorityMedium), nil)

	// wait until consumer has started speaking queued-1
	time.Sleep(30 * time.Millisecond)

	// high-priority interrupts and clears queued-2
	b.Send(textEv("urgent", processor.PriorityHigh), nil)

	// release the block
	released = true
	close(block)
	_ = released

	waitSpoken(t, blocking.mock, 2)

	spoken := blocking.mock.getSpoken()
	// "queued-1" was spoken before interrupt; "urgent" after; "queued-2" dropped
	last := spoken[len(spoken)-1]
	if last != "urgent" {
		t.Errorf("last spoken should be 'urgent', got %q (all: %v)", last, spoken)
	}
	for _, s := range spoken {
		if s == "queued-2" {
			t.Errorf("'queued-2' should have been dropped by PriorityHigh, got: %v", spoken)
		}
	}
}

func TestBridge_LowPriorityDroppedWhenQueueNotEmpty(t *testing.T) {
	block := make(chan struct{})
	blocking := &blockingSpeaker{block: block, mock: &mockSpeaker{}}
	b := tts.NewBridge(blocking, nil)
	defer b.Close()

	b.Send(taskStart(), nil)
	b.Send(textEv("first", processor.PriorityMedium), nil)
	// queue is now non-empty (consumer is blocked) — low priority should be dropped
	b.Send(textEv("low", processor.PriorityLow), nil)

	close(block)
	waitSpoken(t, blocking.mock, 1)

	time.Sleep(30 * time.Millisecond)
	spoken := blocking.mock.getSpoken()
	for _, s := range spoken {
		if s == "low" {
			t.Errorf("low-priority item should have been dropped, got: %v", spoken)
		}
	}
}

func TestBridge_LowPrioritySpokenWhenQueueEmpty(t *testing.T) {
	mock := &mockSpeaker{}
	b := tts.NewBridge(mock, nil)
	defer b.Close()

	b.Send(taskStart(), nil)
	b.Send(textEv("low", processor.PriorityLow), nil)
	waitSpoken(t, mock, 1)

	spoken := mock.getSpoken()
	if len(spoken) != 1 || spoken[0] != "low" {
		t.Errorf("low-priority item should be spoken when queue is empty, got: %v", spoken)
	}
}

func TestBridge_EmptyTextSkipped(t *testing.T) {
	mock := &mockSpeaker{}
	b := tts.NewBridge(mock, nil)
	defer b.Close()

	b.Send(processor.Event{Kind: processor.EventText, Text: "", Priority: processor.PriorityMedium}, nil)
	time.Sleep(30 * time.Millisecond)

	if len(mock.getSpoken()) != 0 {
		t.Errorf("empty text should not be sent to speaker, got: %v", mock.getSpoken())
	}
}

func TestBridge_KindFilter(t *testing.T) {
	mock := &mockSpeaker{}
	// only speak EventError and EventToolUse
	b := tts.NewBridge(mock, []processor.EventKind{processor.EventError, processor.EventToolUse})
	defer b.Close()

	b.Send(taskStart(), nil)
	b.Send(processor.Event{Kind: processor.EventText, Text: "ignored", Priority: processor.PriorityMedium}, nil)
	b.Send(processor.Event{Kind: processor.EventError, Text: "error!", Priority: processor.PriorityHigh}, nil)
	b.Send(processor.Event{Kind: processor.EventSpinner, Text: "spinning", Priority: processor.PriorityMedium}, nil)

	waitSpoken(t, mock, 1)
	time.Sleep(30 * time.Millisecond)

	spoken := mock.getSpoken()
	if len(spoken) != 1 || spoken[0] != "error!" {
		t.Errorf("only EventError should have been spoken, got: %v", spoken)
	}
}

func TestBridge_CloseIsIdempotent(t *testing.T) {
	mock := &mockSpeaker{}
	b := tts.NewBridge(mock, nil)
	if err := b.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := b.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

// blockingSpeaker wraps mockSpeaker and blocks in Wait until the block channel is closed.
type blockingSpeaker struct {
	block <-chan struct{}
	mock  *mockSpeaker
}

func (b *blockingSpeaker) Speak(text string, p processor.Priority) error {
	return b.mock.Speak(text, p)
}
func (b *blockingSpeaker) Wait() error {
	<-b.block
	return nil
}
func (b *blockingSpeaker) Stop() error          { return b.mock.Stop() }
func (b *blockingSpeaker) SetRate(r int) error  { return b.mock.SetRate(r) }
func (b *blockingSpeaker) SetVolume(v int) error { return b.mock.SetVolume(v) }
func (b *blockingSpeaker) Close() error         { return b.mock.Close() }
