// Package processor detects semantic events from a cleaned byte stream.
package processor

import "time"

// EventKind classifies the semantic type of a terminal event.
type EventKind int

const (
	EventText    EventKind = iota // Conversational text from the agent
	EventCode                     // Code block start/end
	EventToolUse                  // Agent is invoking a tool
	EventToolResult               // Tool result received
	EventError                    // Error output
	EventTaskStart                // Agent started a new response
	EventTaskEnd                  // Agent finished responding
	EventSpinner                  // In-progress spinner (suppress in SR mode)
	EventClearScreen              // Clear screen sequence
)

// Priority controls TTS queue behavior.
type Priority int

const (
	PriorityLow    Priority = iota
	PriorityMedium          // enqueue
	PriorityHigh            // interrupt current speech
)

// Event is a semantic unit emitted by the Processor.
type Event struct {
	Kind      EventKind
	Text      string
	Priority  Priority
	Timestamp time.Time
}

// Processor is a stub for phase 2. It will parse a cleaned byte stream
// and emit Events for the renderer and TTS bridge.
type Processor struct{}

// New creates a Processor with default configuration.
func New() *Processor {
	return &Processor{}
}

// Process converts cleaned bytes into a slice of Events.
// Phase 2 will implement full heuristic detection.
func (p *Processor) Process(b []byte) []Event {
	if len(b) == 0 {
		return nil
	}
	return []Event{{
		Kind:      EventText,
		Text:      string(b),
		Priority:  PriorityMedium,
		Timestamp: time.Now(),
	}}
}
