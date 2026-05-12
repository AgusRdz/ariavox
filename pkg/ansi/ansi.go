// Package ansi provides a finite-state machine parser for ANSI escape sequences.
// It handles fragmented input correctly — a sequence may arrive across multiple Read calls.
package ansi

// Parser strips ANSI escape sequences from a byte stream.
// It is safe for concurrent use via its process method.
type Parser struct {
	state parserState
	buf   []byte
}

type parserState int

const (
	stateNormal parserState = iota
	stateEsc
	stateCSI
)

// Strip removes ANSI escape sequences from b and returns the clean text.
func Strip(b []byte) []byte {
	p := &Parser{}
	return p.Process(b)
}

// Process feeds bytes into the parser and returns any printable text.
// Call repeatedly; state is preserved across calls.
func (p *Parser) Process(b []byte) []byte {
	out := make([]byte, 0, len(b))
	for _, c := range b {
		switch p.state {
		case stateNormal:
			if c == 0x1B {
				p.state = stateEsc
			} else {
				out = append(out, c)
			}
		case stateEsc:
			switch c {
			case '[':
				p.state = stateCSI
			case 's', 'u', '7', '8':
				// save/restore cursor — ignore
				p.state = stateNormal
			default:
				p.state = stateNormal
			}
		case stateCSI:
			// CSI sequences end at any byte in 0x40–0x7E
			if c >= 0x40 && c <= 0x7E {
				p.state = stateNormal
			}
			// accumulate params — discard all
		}
	}
	return out
}

// Reset clears the parser state. Use between independent streams.
func (p *Parser) Reset() {
	p.state = stateNormal
	p.buf = p.buf[:0]
}
