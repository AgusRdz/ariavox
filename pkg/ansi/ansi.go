// Package ansi provides a finite-state machine parser for ANSI escape sequences.
// It handles fragmented input correctly — a sequence may arrive across multiple Read calls.
package ansi

// Parser strips ANSI/VT escape sequences from a byte stream.
// It is NOT safe for concurrent use; use one Parser per goroutine.
type Parser struct {
	state parserState
}

type parserState int

const (
	stateNormal parserState = iota
	stateEsc                // received 0x1B
	stateCSI                // received 0x1B [
	stateOSC                // received 0x1B ] — OSC (title set, etc.)
	stateESCSeq             // received 0x1B followed by non-[ non-] char (e.g. 0x1B P, 0x1B _)
)

// Strip removes ANSI/VT escape sequences from b and returns the clean text.
func Strip(b []byte) []byte {
	p := &Parser{}
	return p.Process(b)
}

// Process feeds bytes into the parser and returns printable text only.
// State is preserved across calls so sequences split across reads are handled correctly.
//
// Carriage return (\r) is treated as "erase current line" — everything written
// to the current line so far is discarded, because Claude Code uses \r to
// overwrite spinner lines in place. This prevents the spinner text from being
// concatenated with the response text.
func (p *Parser) Process(b []byte) []byte {
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		c := b[i]
		switch p.state {
		case stateNormal:
			switch c {
			case 0x1B:
				p.state = stateEsc
			case 0x07, 0x0E, 0x0F, 0x0C:
				// BEL, SO, SI, FF — discard
			case '\r':
				// Carriage return: insert a newline so the processor sees
				// each overwritten "subline" as a separate line. The last
				// subline (the real content) wins because \r-overwrite lines
				// are classified independently — spinners are suppressed,
				// ⏺ response lines are kept.
				out = append(out, '\n')
			default:
				out = append(out, c)
			}

		case stateEsc:
			switch c {
			case '[':
				p.state = stateCSI
			case ']':
				p.state = stateOSC
			case 'P', 'X', '^', '_':
				// DCS, SOS, PM, APC — terminated by ST (ESC \)
				p.state = stateESCSeq
			case '\\':
				// ST — string terminator, can arrive after stateESCSeq
				p.state = stateNormal
			default:
				// ESC + single char (e.g. ESC M, ESC =, ESC >) — discard both
				p.state = stateNormal
			}

		case stateCSI:
			// CSI sequences end at a "final byte" in range 0x40–0x7E
			if c >= 0x40 && c <= 0x7E {
				p.state = stateNormal
				// Cursor-forward (C) and cursor-position (H/f) imply whitespace
				// between words when Claude streams tokens. Insert a space so
				// "quétepuedoayudar" doesn't collapse into a single token.
				if c == 'C' || c == 'H' || c == 'f' {
					out = append(out, ' ')
				}
			}
			// all CSI param/intermediate bytes discarded

		case stateOSC:
			// OSC ends at BEL (0x07) or ST (ESC \)
			if c == 0x07 {
				p.state = stateNormal
			} else if c == 0x1B {
				// next byte should be '\' (ST); consume it
				if i+1 < len(b) && b[i+1] == '\\' {
					i++
				}
				p.state = stateNormal
			}
			// all OSC content discarded

		case stateESCSeq:
			// DCS/SOS/PM/APC end at ST (ESC \)
			if c == 0x1B {
				if i+1 < len(b) && b[i+1] == '\\' {
					i++
				}
				p.state = stateNormal
			}
		}
	}
	return out
}

// Reset clears the parser state. Use between independent streams.
func (p *Parser) Reset() {
	p.state = stateNormal
}
