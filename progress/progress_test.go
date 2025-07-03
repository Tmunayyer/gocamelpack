package progress

import (
	"bytes"
	"testing"
)

func TestNoOpReporter(t *testing.T) {
	reporter := NewNoOpReporter()
	
	// Test all methods don't panic and return expected values
	reporter.SetTotal(10)
	reporter.Increment()
	reporter.IncrementBy(5)
	reporter.SetCurrent(8)
	reporter.SetMessage("test message")
	reporter.Finish()
	
	if reporter.IsComplete() {
		t.Error("NoOpReporter.IsComplete() should always return false")
	}
	
	if reporter.Current() != 0 {
		t.Errorf("NoOpReporter.Current() = %d, want 0", reporter.Current())
	}
	
	if reporter.Total() != 0 {
		t.Errorf("NoOpReporter.Total() = %d, want 0", reporter.Total())
	}
}

func TestProgressState_SetTotal(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  int
	}{
		{"positive value", 10, 10},
		{"zero value", 0, 0},
		{"negative value", -5, 0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgressState(&bytes.Buffer{})
			p.SetTotal(tt.input)
			if got := p.Total(); got != tt.want {
				t.Errorf("SetTotal(%d): got %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestProgressState_SetCurrent(t *testing.T) {
	tests := []struct {
		name    string
		total   int
		current int
		want    int
	}{
		{"within bounds", 10, 5, 5},
		{"at total", 10, 10, 10},
		{"exceeds total", 10, 15, 10},
		{"negative value", 10, -3, 0},
		{"zero total, any current", 0, 5, 5}, // When total is 0, current can be anything
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgressState(&bytes.Buffer{})
			p.SetTotal(tt.total)
			p.SetCurrent(tt.current)
			if got := p.Current(); got != tt.want {
				t.Errorf("SetCurrent(%d) with total %d: got %d, want %d", tt.current, tt.total, got, tt.want)
			}
		})
	}
}

func TestProgressState_Increment(t *testing.T) {
	p := NewProgressState(&bytes.Buffer{})
	p.SetTotal(10)
	
	// Test increment from 0
	if got := p.Current(); got != 0 {
		t.Errorf("Initial current: got %d, want 0", got)
	}
	
	p.Increment()
	if got := p.Current(); got != 1 {
		t.Errorf("After Increment(): got %d, want 1", got)
	}
	
	// Test multiple increments
	p.Increment()
	p.Increment()
	if got := p.Current(); got != 3 {
		t.Errorf("After 3 increments: got %d, want 3", got)
	}
	
	// Test increment beyond total
	p.SetCurrent(9)
	p.Increment()
	if got := p.Current(); got != 10 {
		t.Errorf("At total after increment: got %d, want 10", got)
	}
	
	p.Increment() // Should be capped at total
	if got := p.Current(); got != 10 {
		t.Errorf("Beyond total after increment: got %d, want 10", got)
	}
}

func TestProgressState_IncrementBy(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		initial  int
		amount   int
		expected int
	}{
		{"normal increment", 10, 2, 3, 5},
		{"increment to total", 10, 7, 3, 10},
		{"increment beyond total", 10, 8, 5, 10},
		{"zero increment", 10, 5, 0, 5},
		{"negative increment", 10, 5, -2, 5}, // Should not change
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgressState(&bytes.Buffer{})
			p.SetTotal(tt.total)
			p.SetCurrent(tt.initial)
			p.IncrementBy(tt.amount)
			if got := p.Current(); got != tt.expected {
				t.Errorf("IncrementBy(%d) from %d: got %d, want %d", tt.amount, tt.initial, got, tt.expected)
			}
		})
	}
}

func TestProgressState_SetMessage(t *testing.T) {
	p := NewProgressState(&bytes.Buffer{})
	
	testMessage := "Processing file.txt"
	p.SetMessage(testMessage)
	
	if got := p.Message(); got != testMessage {
		t.Errorf("SetMessage(): got %q, want %q", got, testMessage)
	}
	
	// Test empty message
	p.SetMessage("")
	if got := p.Message(); got != "" {
		t.Errorf("SetMessage(empty): got %q, want empty string", got)
	}
}

func TestProgressState_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		current  int
		expected bool
	}{
		{"not complete", 10, 5, false},
		{"exactly complete", 10, 10, true}, // current >= total means complete
		{"over complete", 10, 12, true},    // current >= total means complete
		{"zero total not complete", 0, 5, false}, // Special case: if total is 0, never complete
		{"zero both not complete", 0, 0, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgressState(&bytes.Buffer{})
			p.SetTotal(tt.total)
			p.SetCurrent(tt.current)
			if got := p.IsComplete(); got != tt.expected {
				t.Errorf("IsComplete() with total=%d, current=%d: got %v, want %v", tt.total, tt.current, got, tt.expected)
			}
		})
	}
}

func TestProgressState_Percentage(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		current  int
		expected int
	}{
		{"zero percent", 10, 0, 0},
		{"fifty percent", 10, 5, 50},
		{"hundred percent", 10, 10, 100},
		{"over hundred percent", 10, 12, 120},
		{"zero total", 0, 5, 0}, // Should not divide by zero
		{"rounding down", 3, 1, 33}, // 33.33% rounds down to 33
		{"rounding up", 3, 2, 66},   // 66.66% rounds down to 66
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgressState(&bytes.Buffer{})
			p.SetTotal(tt.total)
			p.SetCurrent(tt.current)
			if got := p.Percentage(); got != tt.expected {
				t.Errorf("Percentage() with total=%d, current=%d: got %d%%, want %d%%", tt.total, tt.current, got, tt.expected)
			}
		})
	}
}

func TestProgressState_String(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		current  int
		expected string
	}{
		{"with total", 10, 3, "3/10 (30%)"},
		{"complete", 10, 10, "10/10 (100%)"},
		{"over complete", 10, 12, "10/10 (120%)"}, // Current gets capped but percentage calculated from actualCurrent
		{"zero total", 0, 5, "5 items processed"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgressState(&bytes.Buffer{})
			p.SetTotal(tt.total)
			p.SetCurrent(tt.current)
			if got := p.String(); got != tt.expected {
				t.Errorf("String() with total=%d, current=%d: got %q, want %q", tt.total, tt.current, got, tt.expected)
			}
		})
	}
}

func TestProgressState_NewProgressState(t *testing.T) {
	buf := &bytes.Buffer{}
	p := NewProgressState(buf)
	
	if p == nil {
		t.Error("NewProgressState() returned nil")
	}
	
	if p.writer != buf {
		t.Error("NewProgressState() did not set writer correctly")
	}
	
	if p.Current() != 0 {
		t.Errorf("NewProgressState() current: got %d, want 0", p.Current())
	}
	
	if p.Total() != 0 {
		t.Errorf("NewProgressState() total: got %d, want 0", p.Total())
	}
	
	if p.Message() != "" {
		t.Errorf("NewProgressState() message: got %q, want empty", p.Message())
	}
}