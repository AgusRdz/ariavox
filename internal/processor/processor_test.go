package processor_test

import (
	"os"
	"testing"

	"github.com/AgusRdz/ariavox/internal/processor"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("../../testdata/" + name)
	if err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return b
}

func TestSpinnerLines(t *testing.T) {
	p := processor.New()
	events := p.Process(fixture(t, "claude_spinner.txt"))

	for _, e := range events {
		if e.Kind != processor.EventSpinner {
			t.Errorf("expected EventSpinner, got %v (text=%q)", e.Kind, e.Text)
		}
	}
	if len(events) == 0 {
		t.Error("expected spinner events, got none")
	}
}

func TestCodeBlock(t *testing.T) {
	p := processor.New()
	events := p.Process(fixture(t, "claude_code_block.txt"))

	var kinds []processor.EventKind
	for _, e := range events {
		kinds = append(kinds, e.Kind)
	}

	hasCode := false
	hasText := false
	hasTaskStart := false
	for _, k := range kinds {
		switch k {
		case processor.EventCode:
			hasCode = true
		case processor.EventText:
			hasText = true
		case processor.EventTaskStart:
			hasTaskStart = true
		}
	}

	if !hasCode {
		t.Error("expected EventCode events from code block fixture")
	}
	if !hasText {
		t.Error("expected EventText events from code block fixture")
	}
	if !hasTaskStart {
		t.Error("expected EventTaskStart at beginning")
	}
}

func TestToolUse(t *testing.T) {
	p := processor.New()
	events := p.Process(fixture(t, "claude_tool_use.txt"))

	hasToolUse := false
	hasToolResult := false
	for _, e := range events {
		if e.Kind == processor.EventToolUse {
			hasToolUse = true
		}
		if e.Kind == processor.EventToolResult {
			hasToolResult = true
		}
	}

	if !hasToolUse {
		t.Error("expected EventToolUse events")
	}
	if !hasToolResult {
		t.Error("expected EventToolResult events")
	}
}

func TestErrorDetection(t *testing.T) {
	p := processor.New()
	events := p.Process(fixture(t, "claude_error.txt"))

	hasError := false
	for _, e := range events {
		if e.Kind == processor.EventError {
			hasError = true
			if e.Priority != processor.PriorityHigh {
				t.Errorf("error event should have PriorityHigh, got %v", e.Priority)
			}
		}
	}
	if !hasError {
		t.Error("expected EventError events")
	}
}

func TestLongResponse(t *testing.T) {
	p := processor.New()
	events := p.Process(fixture(t, "claude_long_response.txt"))

	taskStarts := 0
	for _, e := range events {
		if e.Kind == processor.EventTaskStart {
			taskStarts++
		}
	}
	if taskStarts != 1 {
		t.Errorf("expected exactly 1 EventTaskStart, got %d", taskStarts)
	}
}

func TestTaskStartOnlyOnce(t *testing.T) {
	p := processor.New()

	// feed multiple chunks; TaskStart should appear only once
	chunks := []string{
		"First line\n",
		"Second line\n",
		"Third line\n",
	}

	var all []processor.Event
	for _, chunk := range chunks {
		all = append(all, p.Process([]byte(chunk))...)
	}

	taskStarts := 0
	for _, e := range all {
		if e.Kind == processor.EventTaskStart {
			taskStarts++
		}
	}
	if taskStarts != 1 {
		t.Errorf("expected 1 TaskStart across chunks, got %d", taskStarts)
	}
}

func TestCodeBlockState(t *testing.T) {
	p := processor.New()
	input := "```go\nfunc main() {}\n```\n"
	events := p.Process([]byte(input))

	codeCount := 0
	for _, e := range events {
		if e.Kind == processor.EventCode {
			codeCount++
		}
	}
	// opening fence + body line + closing fence = 3 code events
	if codeCount != 3 {
		t.Errorf("expected 3 EventCode events, got %d", codeCount)
	}
}

func TestFragmentedInput(t *testing.T) {
	p := processor.New()

	// line split across two reads
	ev1 := p.Process([]byte("hello "))
	ev2 := p.Process([]byte("world\n"))

	all := append(ev1, ev2...)
	if len(all) == 0 {
		t.Fatal("expected events after full line received")
	}

	found := false
	for _, e := range all {
		if e.Kind == processor.EventText && e.Text == "hello world" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected EventText with 'hello world', got %+v", all)
	}
}

func TestEmptyInput(t *testing.T) {
	p := processor.New()
	events := p.Process(nil)
	if events != nil {
		t.Errorf("expected nil for empty input, got %v", events)
	}
}

func TestReset(t *testing.T) {
	p := processor.New()
	p.Process([]byte("some text\n"))
	p.Reset()

	// after reset, TaskStart should fire again
	events := p.Process([]byte("new session\n"))
	hasTaskStart := false
	for _, e := range events {
		if e.Kind == processor.EventTaskStart {
			hasTaskStart = true
		}
	}
	if !hasTaskStart {
		t.Error("expected TaskStart after Reset")
	}
}
