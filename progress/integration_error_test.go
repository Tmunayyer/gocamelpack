package progress

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

// MockFailingProgressReporter simulates a progress reporter that fails partway through
type MockFailingProgressReporter struct {
	*ProgressBar
	failAt int
	calls  int
}

func NewMockFailingProgressReporter(writer *bytes.Buffer, failAt int) *MockFailingProgressReporter {
	return &MockFailingProgressReporter{
		ProgressBar: NewProgressBar(writer, 30),
		failAt:      failAt,
		calls:       0,
	}
}

func (m *MockFailingProgressReporter) SetCurrent(current int) {
	m.calls++
	if m.calls >= m.failAt {
		m.ProgressBar.SetError(errors.New("simulated failure"))
		return
	}
	m.ProgressBar.SetCurrent(current)
}

func TestProgressBar_IntegrationWithError(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewMockFailingProgressReporter(buf, 3) // Fail on 3rd call
	
	// Simulate a typical file operation sequence
	reporter.SetTotal(5)
	reporter.SetMessage("Processing file 1")
	reporter.SetCurrent(1) // Call 1 - should work
	
	reporter.SetMessage("Processing file 2")
	reporter.SetCurrent(2) // Call 2 - should work
	
	reporter.SetMessage("Processing file 3")
	reporter.SetCurrent(3) // Call 3 - should trigger error
	
	// These should be ignored after error
	reporter.SetMessage("Processing file 4")
	reporter.SetCurrent(4)
	reporter.Finish()
	
	output := buf.String()
	
	// Should contain error indicator
	if !reporter.IsErrored() {
		t.Error("Expected reporter to be in error state")
	}
	
	// Should show error in output
	if output == "" {
		t.Error("Expected some output from error state")
	}
	
	// Should not have reached final state
	if reporter.Current() != 2 { // Should be stuck at last successful update
		t.Errorf("Expected current to be 2, got %d", reporter.Current())
	}
}

func TestProgressBar_RecoveryAfterError(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewProgressBar(buf, 20)
	bar.SetTotal(10)
	bar.SetCurrent(5)
	
	// Trigger error
	testErr := errors.New("test failure")
	bar.SetError(testErr)
	
	// Clear buffer for new test
	buf.Reset()
	
	// Create new bar for recovery (simulating retry)
	recoveryBar := NewProgressBar(buf, 20)
	recoveryBar.SetTotal(10)
	recoveryBar.SetMessage("Retrying operations")
	recoveryBar.SetCurrent(5) // Start where we left off
	
	for i := 6; i <= 10; i++ {
		recoveryBar.SetMessage(fmt.Sprintf("Retry processing item %d", i))
		recoveryBar.SetCurrent(i)
	}
	recoveryBar.Finish()
	
	output := buf.String()
	
	// Should show successful completion
	if !recoveryBar.IsComplete() {
		t.Error("Expected recovery bar to be complete")
	}
	
	if recoveryBar.IsErrored() {
		t.Error("Expected recovery bar to not be errored")
	}
	
	// Should contain success indicator
	if !bytes.Contains([]byte(output), []byte("✓")) {
		t.Error("Expected success checkmark in recovery output")
	}
}

func TestProgressBar_ErrorAtDifferentStages(t *testing.T) {
	testCases := []struct {
		name      string
		total     int
		errorAt   int
		expectMsg string
	}{
		{"Error at start", 10, 0, "0/10"},
		{"Error at middle", 10, 5, "5/10"},
		{"Error near end", 10, 9, "9/10"},
		{"Error at completion", 10, 10, "10/10"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			bar := NewProgressBar(buf, 20)
			bar.SetTotal(tc.total)
			bar.SetCurrent(tc.errorAt)
			
			testErr := fmt.Errorf("failure at stage %d", tc.errorAt)
			bar.SetError(testErr)
			
			output := buf.String()
			
			// Should contain expected progress
			if !bytes.Contains([]byte(output), []byte(tc.expectMsg)) {
				t.Errorf("Expected %q in output, got: %q", tc.expectMsg, output)
			}
			
			// Should contain error indicator
			if !bytes.Contains([]byte(output), []byte("✗")) {
				t.Error("Expected error indicator in output")
			}
			
			// Should be errored
			if !bar.IsErrored() {
				t.Error("Expected bar to be errored")
			}
		})
	}
}