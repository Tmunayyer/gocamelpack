package progress

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestProgressBar_SetError(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewProgressBar(buf, 20)
	bar.SetTotal(10)
	bar.SetCurrent(5) // 50% complete

	testErr := errors.New("operation failed")
	bar.SetError(testErr)

	output := buf.String()

	// Should contain error indicator
	if !strings.Contains(output, "✗") {
		t.Error("Expected error mark (✗) in output")
	}

	// Should contain error message
	if !strings.Contains(output, "Error: operation failed") {
		t.Error("Expected error message in output")
	}

	// Should show partial progress
	if !strings.Contains(output, "5/10") {
		t.Error("Expected progress stats in error output")
	}

	// Should be marked as errored
	if !bar.IsErrored() {
		t.Error("Expected progress bar to be marked as errored")
	}

	// Should be marked as finished
	if !bar.finished {
		t.Error("Expected progress bar to be finished after error")
	}
}

func TestProgressBar_SetError_NilError(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewProgressBar(buf, 20)
	bar.SetTotal(10)
	bar.SetCurrent(3)

	bar.SetError(nil)

	output := buf.String()

	// Should contain error indicator
	if !strings.Contains(output, "✗") {
		t.Error("Expected error mark (✗) in output")
	}

	// Should not contain error message when error is nil
	if strings.Contains(output, "Error:") {
		t.Error("Should not contain error message when error is nil")
	}

	// Should be marked as errored
	if !bar.IsErrored() {
		t.Error("Expected progress bar to be marked as errored")
	}
}

func TestProgressBar_SetError_WithMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewProgressBar(buf, 20)
	bar.SetTotal(5)
	bar.SetCurrent(2)
	bar.SetMessage("Processing file.txt")

	testErr := errors.New("permission denied")
	bar.SetError(testErr)

	output := buf.String()

	// Should contain current message
	if !strings.Contains(output, "Processing file.txt") {
		t.Error("Expected current message in error output")
	}

	// Should contain error details
	if !strings.Contains(output, "Error: permission denied") {
		t.Error("Expected specific error message in output")
	}
}

func TestProgressBar_SetError_AfterFinished(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewProgressBar(buf, 20)
	bar.SetTotal(10)
	bar.SetCurrent(10)
	bar.Finish()

	// Clear buffer to test no additional output
	buf.Reset()

	testErr := errors.New("late error")
	bar.SetError(testErr)

	// Should not output anything since already finished
	if buf.Len() > 0 {
		t.Errorf("Expected no output after setting error on finished bar, got: %q", buf.String())
	}

	// Should not be marked as errored since it was already finished
	if bar.IsErrored() {
		t.Error("Should not mark finished bar as errored")
	}
}

func TestProgressBar_IsErrored_InitialState(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewProgressBar(buf, 20)

	// Should not be errored initially
	if bar.IsErrored() {
		t.Error("New progress bar should not be errored initially")
	}
}

func TestProgressBar_SetError_ZeroProgress(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewProgressBar(buf, 20)
	bar.SetTotal(10)
	// Don't set current, should be 0

	testErr := errors.New("early failure")
	bar.SetError(testErr)

	output := buf.String()

	// Should show 0 progress
	if !strings.Contains(output, "0/10") {
		t.Error("Expected 0/10 progress in error output")
	}

	// Should still have progress bar structure
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("Expected progress bar brackets in error output")
	}
}

func TestProgressBar_SetError_FullProgress(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewProgressBar(buf, 20)
	bar.SetTotal(10)
	bar.SetCurrent(10) // Complete

	testErr := errors.New("error at completion")
	bar.SetError(testErr)

	output := buf.String()

	// Should show full progress
	if !strings.Contains(output, "10/10") {
		t.Error("Expected 10/10 progress in error output")
	}

	// Should still have error indicator
	if !strings.Contains(output, "✗") {
		t.Error("Expected error mark even at full progress")
	}
}

func TestProgressBar_ErrorWithUpdate_NoEffect(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewProgressBar(buf, 20)
	bar.SetTotal(10)
	bar.SetCurrent(5)

	testErr := errors.New("test error")
	bar.SetError(testErr)

	// Clear buffer to check no additional output
	buf.Reset()

	// Try to update after error - should have no effect
	bar.SetCurrent(7)
	bar.SetMessage("new message")
	bar.Update()

	// Should not produce any additional output
	if buf.Len() > 0 {
		t.Errorf("Expected no output after updating errored bar, got: %q", buf.String())
	}

	// Current should remain at error state (5)
	if bar.Current() != 5 {
		t.Errorf("Expected current to remain 5 after error, got %d", bar.Current())
	}
}