// Package processor detects semantic events from a cleaned byte stream.
package processor

import (
	"bytes"
	"strings"
	"time"
	"unicode/utf8"
)

// EventKind classifies the semantic type of a terminal event.
type EventKind int

const (
	EventText        EventKind = iota // Conversational text from the agent
	EventCode                         // Code block start/end
	EventToolUse                      // Agent is invoking a tool
	EventToolResult                   // Tool result received
	EventError                        // Error output
	EventTaskStart                    // Agent started a new response
	EventTaskEnd                      // Agent finished responding
	EventSpinner                      // In-progress spinner (suppress in SR mode)
	EventClearScreen                  // Clear screen sequence
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

// Processor parses a cleaned byte stream and emits semantic Events.
// It is stateful — a single Processor instance must be used for a continuous stream.
type Processor struct {
	inCodeBlock   bool   // inside a ``` fence
	codeLang      string // language tag on the opening fence, if any
	taskStarted   bool   // whether TaskStart has been emitted for the current response
	inResponse    bool   // true after ⏺ is seen, false after response ends
	lineBuf       []byte // accumulates incomplete lines across Read calls
}

// New creates a Processor with default configuration.
func New() *Processor {
	return &Processor{}
}

// Process converts cleaned bytes into a slice of Events.
// Input must have ANSI sequences already stripped.
func (p *Processor) Process(b []byte) []Event {
	if len(b) == 0 {
		return nil
	}

	// append to any buffered partial line
	data := append(p.lineBuf, b...)
	p.lineBuf = nil

	var events []Event

	for len(data) > 0 {
		idx := bytes.IndexByte(data, '\n')
		var line []byte
		if idx == -1 {
			// no newline yet — buffer and wait for more data
			p.lineBuf = make([]byte, len(data))
			copy(p.lineBuf, data)
			break
		}
		line = data[:idx]
		data = data[idx+1:]

		// strip trailing CR
		line = bytes.TrimRight(line, "\r")

		evs := p.classifyLine(string(line))
		events = append(events, evs...)
	}

	return events
}

// classifyLine maps a single cleaned line to zero or more Events.
func (p *Processor) classifyLine(line string) []Event {
	// clear-screen sentinel
	if line == "\x0c" {
		return []Event{ev(EventClearScreen, "", PriorityLow)}
	}

	// inside a code block: everything is code until closing fence
	if p.inCodeBlock {
		if isCodeFence(line) {
			p.inCodeBlock = false
			p.codeLang = ""
			return []Event{ev(EventCode, line, PriorityLow)}
		}
		return []Event{ev(EventCode, line, PriorityLow)}
	}

	// spinner lines — also end response mode (spinner after response = Claude finished)
	if isSpinnerLine(line) {
		if p.inResponse {
			p.inResponse = false
		}
		return []Event{ev(EventSpinner, line, PriorityHigh)}
	}

	// UI chrome — box drawing, prompt, status bar — also ends response mode
	if isUIChrome(line) {
		p.inResponse = false
		return nil
	}

	// ⏺ marks the start of a Claude response line — enter response mode.
	if strings.HasPrefix(line, "⏺") {
		content := strings.TrimSpace(strings.TrimPrefix(line, "⏺"))
		p.inResponse = true
		if content == "" {
			return nil
		}
		return p.emitWithTaskStart(ev(EventText, content, PriorityMedium))
	}

	// tool use: lines starting with ● — exit response mode
	if strings.HasPrefix(line, "●") {
		p.inResponse = false
		return p.emitWithTaskStart(ev(EventToolUse, line, PriorityHigh))
	}

	// tool result indent lines: ⎿
	if strings.Contains(line, "⎿") {
		p.inResponse = false
		kind := EventToolResult
		if strings.Contains(line, "Error:") || strings.Contains(line, "error:") {
			kind = EventError
			return p.emitWithTaskStart(ev(kind, line, PriorityHigh))
		}
		return p.emitWithTaskStart(ev(kind, line, PriorityMedium))
	}

	// standalone error lines
	if isErrorLine(line) {
		return p.emitWithTaskStart(ev(EventError, line, PriorityHigh))
	}

	// empty lines
	if strings.TrimSpace(line) == "" {
		return nil
	}

	// Only emit text events while inside a Claude response.
	if !p.inResponse {
		return nil
	}

	// opening code fence inside a response
	if isCodeFence(line) {
		p.inCodeBlock = true
		p.codeLang = codeFenceLang(line)
		return []Event{ev(EventCode, line, PriorityLow)}
	}

	return p.emitWithTaskStart(ev(EventText, line, PriorityMedium))
}

// emitWithTaskStart prepends a TaskStart event the first time a non-spinner
// event is produced in the current response.
func (p *Processor) emitWithTaskStart(e Event) []Event {
	if !p.taskStarted {
		p.taskStarted = true
		return []Event{ev(EventTaskStart, "", PriorityHigh), e}
	}
	return []Event{e}
}

// Reset clears all state. Call when the child process exits or a new session begins.
func (p *Processor) Reset() {
	p.inCodeBlock = false
	p.codeLang = ""
	p.taskStarted = false
	p.inResponse = false
	p.lineBuf = nil
}

// --- helpers ---

func ev(kind EventKind, text string, priority Priority) Event {
	return Event{Kind: kind, Text: text, Priority: priority, Timestamp: time.Now()}
}

// spinnerRunes includes Braille spinners and Claude Code's star/dot spinners.
// ⏺ (U+23FA) is NOT included — it is Claude's response bullet, not a spinner.
var spinnerRunes = []rune{
	'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏', // Braille
	'✳', '✶', '✻', '✢', '·', '⠂', '⠐',                  // Claude Code UI spinners
}

func isSpinnerLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) == 0 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(trimmed)
	for _, s := range spinnerRunes {
		if r == s {
			return true
		}
	}
	return false
}

func isCodeFence(line string) bool {
	s := strings.TrimLeft(line, " \t")
	return strings.HasPrefix(s, "```")
}

func codeFenceLang(line string) string {
	s := strings.TrimLeft(line, " \t")
	s = strings.TrimPrefix(s, "```")
	return strings.TrimSpace(s)
}

// errorPrefixes are patterns that indicate error output from tools.
var errorPrefixes = []string{
	"Error:",
	"error:",
	"ERROR:",
	"Fatal:",
	"fatal:",
	"panic:",
}

func isErrorLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	for _, prefix := range errorPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

// uiChromeRunes are box-drawing and decoration characters used exclusively
// in the Claude Code UI shell — never in actual agent responses.
var uiChromeRunes = []rune{'─', '╭', '╰', '│', '╮', '╯', '├', '┤', '┬', '┴', '┼', '▎'}

// isUIChrome returns true for lines that are part of the Claude Code shell UI
// rather than agent response content: box borders, the input prompt, status bar.
func isUIChrome(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) == 0 {
		return false
	}
	// Prompt line: ❯ (input prompt)
	if strings.HasPrefix(trimmed, "❯") {
		return true
	}
	// Shortcut hint line
	if strings.HasPrefix(trimmed, "?") && strings.Contains(trimmed, "shortcuts") {
		return true
	}
	// Lines composed entirely of box-drawing characters and spaces
	allChrome := true
	for _, r := range trimmed {
		isChrome := false
		for _, cr := range uiChromeRunes {
			if r == cr {
				isChrome = true
				break
			}
		}
		if !isChrome && r != ' ' {
			allChrome = false
			break
		}
	}
	return allChrome
}
