package progress

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewProgressBar(t *testing.T) {
	buf := &bytes.Buffer{}
	
	tests := []struct {
		name          string
		width         int
		expectedWidth int
	}{
		{"positive width", 20, 20},
		{"zero width", 0, 40}, // Should use default
		{"negative width", -5, 40}, // Should use default
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewProgressBar(buf, tt.width)
			if pb.width != tt.expectedWidth {
				t.Errorf("NewProgressBar(width=%d): got width %d, want %d", tt.width, pb.width, tt.expectedWidth)
			}
			
			if pb.ProgressState == nil {
				t.Error("NewProgressBar(): ProgressState is nil")
			}
			
			if pb.barChar != '█' {
				t.Errorf("NewProgressBar(): barChar = %c, want █", pb.barChar)
			}
			
			if pb.emptyChar != '░' {
				t.Errorf("NewProgressBar(): emptyChar = %c, want ░", pb.emptyChar)
			}
			
			if !pb.showMsg {
				t.Error("NewProgressBar(): showMsg should be true by default")
			}
		})
	}
}

func TestNewSimpleProgressBar(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewSimpleProgressBar(buf)
	
	if pb.width != 40 {
		t.Errorf("NewSimpleProgressBar(): width = %d, want 40", pb.width)
	}
}

func TestProgressBar_SetBarChar(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	
	pb.SetBarChar('=')
	if pb.barChar != '=' {
		t.Errorf("SetBarChar('='): got %c, want =", pb.barChar)
	}
}

func TestProgressBar_SetEmptyChar(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	
	pb.SetEmptyChar('-')
	if pb.emptyChar != '-' {
		t.Errorf("SetEmptyChar('-'): got %c, want -", pb.emptyChar)
	}
}

func TestProgressBar_SetShowMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	
	pb.SetShowMessage(false)
	if pb.showMsg {
		t.Error("SetShowMessage(false): showMsg should be false")
	}
	
	pb.SetShowMessage(true)
	if !pb.showMsg {
		t.Error("SetShowMessage(true): showMsg should be true")
	}
}

func TestProgressBar_SetWidth(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	
	pb.SetWidth(20)
	if pb.width != 20 {
		t.Errorf("SetWidth(20): got %d, want 20", pb.width)
	}
	
	// Test invalid width
	pb.SetWidth(0)
	if pb.width != 20 {
		t.Errorf("SetWidth(0): width should remain 20, got %d", pb.width)
	}
	
	pb.SetWidth(-5)
	if pb.width != 20 {
		t.Errorf("SetWidth(-5): width should remain 20, got %d", pb.width)
	}
}

func TestProgressBar_Render(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	pb.SetBarChar('=')
	pb.SetEmptyChar('-')
	
	tests := []struct {
		name     string
		total    int
		current  int
		message  string
		showMsg  bool
		expected string
	}{
		{
			name:     "empty bar",
			total:    10,
			current:  0,
			expected: "[----------] 0/10 (0%)",
		},
		{
			name:     "half full",
			total:    10,
			current:  5,
			expected: "[=====-----] 5/10 (50%)",
		},
		{
			name:     "full bar",
			total:    10,
			current:  10,
			expected: "[==========] 10/10 (100%)",
		},
		{
			name:     "over full",
			total:    10,
			current:  15,
			expected: "[==========] 10/10 (150%)",
		},
		{
			name:     "with message",
			total:    10,
			current:  3,
			message:  "Processing file.txt",
			showMsg:  true,
			expected: "[===-------] 3/10 (30%) - Processing file.txt",
		},
		{
			name:     "message disabled",
			total:    10,
			current:  3,
			message:  "Processing file.txt",
			showMsg:  false,
			expected: "[===-------] 3/10 (30%)",
		},
		{
			name:     "zero total",
			total:    0,
			current:  5,
			expected: "[----------] 5 items processed",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb.SetTotal(tt.total)
			pb.SetCurrent(tt.current)
			pb.SetMessage(tt.message)
			pb.SetShowMessage(tt.showMsg)
			
			got := pb.Render()
			if got != tt.expected {
				t.Errorf("Render(): got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestProgressBar_RenderAfterFinish(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	
	pb.SetTotal(10)
	pb.SetCurrent(5)
	
	// Should render normally before finish
	rendered := pb.Render()
	if rendered == "" {
		t.Error("Render() before finish: should not be empty")
	}
	
	pb.Finish()
	
	// Should not render after finish
	rendered = pb.Render()
	if rendered != "" {
		t.Errorf("Render() after finish: got %q, want empty string", rendered)
	}
}

func TestProgressBar_Display(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	pb.SetBarChar('=')
	pb.SetEmptyChar('-')
	
	// Clear buffer to only test the Display call
	buf.Reset()
	
	pb.SetTotal(10)
	buf.Reset() // Clear again after SetTotal
	
	pb.SetCurrent(3)
	buf.Reset() // Clear again after SetCurrent
	
	pb.Display()
	
	output := buf.String()
	expected := "\r[===-------] 3/10 (30%)"
	
	if output != expected {
		t.Errorf("Display(): got %q, want %q", output, expected)
	}
}

func TestProgressBar_DisplayAfterFinish(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	
	pb.SetTotal(10)
	pb.SetCurrent(5)
	pb.Finish()
	
	buf.Reset()
	pb.Display()
	
	// Should not write anything after finish
	if buf.Len() > 0 {
		t.Errorf("Display() after finish: wrote %q, should write nothing", buf.String())
	}
}

func TestProgressBar_Finish(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	pb.SetBarChar('=')
	
	pb.SetTotal(10)
	pb.SetCurrent(7)
	pb.SetMessage("Processing complete")
	
	pb.Finish()
	
	output := buf.String()
	
	// Should show completed bar with checkmark
	if !strings.Contains(output, "[==========]") {
		t.Error("Finish(): should show full bar")
	}
	
	if !strings.Contains(output, "✓") {
		t.Error("Finish(): should contain checkmark")
	}
	
	if !strings.Contains(output, "\n") {
		t.Error("Finish(): should end with newline")
	}
	
	if !strings.Contains(output, "Processing complete") {
		t.Error("Finish(): should show final message")
	}
	
	// Test multiple calls to Finish (should not duplicate output)
	buf.Reset()
	pb.Finish()
	if buf.Len() > 0 {
		t.Error("Second call to Finish(): should not produce output")
	}
}

func TestProgressBar_IncrementWithDisplay(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	pb.SetBarChar('=')
	pb.SetEmptyChar('-')
	
	pb.SetTotal(10)
	buf.Reset() // Clear after setup
	
	pb.Increment()
	
	output := buf.String()
	expected := "\r[=---------] 1/10 (10%)"
	
	if output != expected {
		t.Errorf("Increment(): got %q, want %q", output, expected)
	}
}

func TestProgressBar_IncrementByWithDisplay(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	pb.SetBarChar('=')
	pb.SetEmptyChar('-')
	
	pb.SetTotal(10)
	buf.Reset() // Clear after setup
	
	pb.IncrementBy(3)
	
	output := buf.String()
	expected := "\r[===-------] 3/10 (30%)"
	
	if output != expected {
		t.Errorf("IncrementBy(3): got %q, want %q", output, expected)
	}
}

func TestProgressBar_SetCurrentWithDisplay(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	pb.SetBarChar('=')
	pb.SetEmptyChar('-')
	
	pb.SetTotal(10)
	buf.Reset() // Clear after setup
	
	pb.SetCurrent(6)
	
	output := buf.String()
	expected := "\r[======----] 6/10 (60%)"
	
	if output != expected {
		t.Errorf("SetCurrent(6): got %q, want %q", output, expected)
	}
}

func TestProgressBar_SetTotalWithDisplay(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	pb.SetBarChar('=')
	pb.SetEmptyChar('-')
	
	buf.Reset() // Clear initial creation display
	pb.SetCurrent(5)
	buf.Reset() // Clear after SetCurrent
	
	pb.SetTotal(20) // This should trigger display
	
	output := buf.String()
	expected := "\r[==--------] 5/20 (25%)"
	
	if output != expected {
		t.Errorf("SetTotal(20): got %q, want %q", output, expected)
	}
}

func TestProgressBar_SetMessageWithDisplay(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	pb.SetBarChar('=')
	pb.SetEmptyChar('-')
	
	pb.SetTotal(10)
	pb.SetCurrent(4)
	buf.Reset() // Clear after setup
	
	pb.SetMessage("New message")
	
	output := buf.String()
	expected := "\r[====------] 4/10 (40%) - New message"
	
	if output != expected {
		t.Errorf("SetMessage(): got %q, want %q", output, expected)
	}
}

func TestProgressBar_ImplementsProgressReporter(t *testing.T) {
	buf := &bytes.Buffer{}
	pb := NewProgressBar(buf, 10)
	
	// This test ensures ProgressBar implements ProgressReporter interface
	var _ ProgressReporter = pb
	
	// Test that all interface methods work
	pb.SetTotal(10)
	pb.SetCurrent(3)
	pb.Increment()
	pb.IncrementBy(2)
	pb.SetMessage("test")
	
	// Should have current=6, total=10, so should be complete
	if pb.Current() != 6 {
		t.Errorf("Current(): got %d, want 6", pb.Current())
	}
	
	if pb.Total() != 10 {
		t.Errorf("Total(): got %d, want 10", pb.Total())
	}
	
	// Complete the progress to test IsComplete
	pb.SetCurrent(10)
	pb.Finish()
	
	if !pb.IsComplete() {
		t.Error("ProgressBar should be complete after reaching total")
	}
}