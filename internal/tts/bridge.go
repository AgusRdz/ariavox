package tts

import (
	"sync"

	"github.com/AgusRdz/ariavox/internal/processor"
)

// item is a queued speech request.
type item struct {
	text     string
	priority processor.Priority
}

// Bridge sits between the processor and a Speaker.
// It serializes speech requests and enforces priority rules:
//
//   - PriorityHigh  — stops current speech, clears queue, speaks immediately
//   - PriorityMedium — enqueues; spoken after the current item finishes
//   - PriorityLow   — enqueues only when queue is empty (dropped otherwise)
//
// Bridge also applies an event-kind filter so only relevant kinds reach TTS.
type Bridge struct {
	speaker Speaker
	filter  map[processor.EventKind]bool

	mu      sync.Mutex
	queue   []item
	wakeup  chan struct{}
	stopCh  chan struct{}
	stopped bool
	wg      sync.WaitGroup
}

// NewBridge creates a Bridge wrapping speaker.
// filter lists the EventKinds that should be spoken; nil means all kinds.
func NewBridge(speaker Speaker, filter []processor.EventKind) *Bridge {
	b := &Bridge{
		speaker: speaker,
		wakeup:  make(chan struct{}, 1),
		stopCh:  make(chan struct{}),
	}
	if filter != nil {
		b.filter = make(map[processor.EventKind]bool, len(filter))
		for _, k := range filter {
			b.filter[k] = true
		}
	}
	b.wg.Add(1)
	go b.run()
	return b
}

// Send queues or immediately handles an event according to its priority.
// Errors from the Speaker are logged to the provided logf function (may be nil).
func (b *Bridge) Send(e processor.Event, logf func(string, ...any)) {
	if b.filter != nil && !b.filter[e.Kind] {
		return
	}
	text := e.Text
	if text == "" {
		return
	}

	b.mu.Lock()
	if b.stopped {
		b.mu.Unlock()
		return
	}

	switch e.Priority {
	case processor.PriorityHigh:
		// stop current speech and discard pending items
		b.queue = b.queue[:0]
		b.queue = append(b.queue, item{text: text, priority: e.Priority})
		if err := b.speaker.Stop(); err != nil && logf != nil {
			logf("tts: stop: %v", err)
		}

	case processor.PriorityMedium:
		b.queue = append(b.queue, item{text: text, priority: e.Priority})

	case processor.PriorityLow:
		if len(b.queue) == 0 {
			b.queue = append(b.queue, item{text: text, priority: e.Priority})
		}
		// otherwise drop: low-priority items are not worth queuing behind other work
	}
	b.mu.Unlock()

	// notify consumer without blocking
	select {
	case b.wakeup <- struct{}{}:
	default:
	}
}

// Close drains the queue and shuts down the background goroutine.
func (b *Bridge) Close() error {
	b.mu.Lock()
	if b.stopped {
		b.mu.Unlock()
		return nil
	}
	b.stopped = true
	b.mu.Unlock()

	close(b.stopCh)
	b.wg.Wait()
	return b.speaker.Close()
}

// SetRate forwards to the underlying Speaker.
func (b *Bridge) SetRate(rate int) error { return b.speaker.SetRate(rate) }

// SetVolume forwards to the underlying Speaker.
func (b *Bridge) SetVolume(vol int) error { return b.speaker.SetVolume(vol) }

// run is the single background consumer goroutine.
func (b *Bridge) run() {
	defer b.wg.Done()
	for {
		select {
		case <-b.stopCh:
			return
		case <-b.wakeup:
			b.drainQueue()
		}
	}
}

func (b *Bridge) drainQueue() {
	for {
		b.mu.Lock()
		if len(b.queue) == 0 {
			b.mu.Unlock()
			return
		}
		it := b.queue[0]
		b.queue = b.queue[1:]
		b.mu.Unlock()

		if err := b.speaker.Speak(it.text, it.priority); err != nil {
			continue
		}
		// Wait releases naturally when speech finishes, or immediately after
		// Stop() was called by a PriorityHigh Send — either way we loop and
		// check the queue again.
		_ = b.speaker.Wait()
	}
}
