package progress

import (
	"fmt"
	"io"
)

// ProgressReporter defines the interface for reporting progress during file operations.
type ProgressReporter interface {
	// SetTotal sets the total number of items to be processed
	SetTotal(total int)
	
	// Increment increases the current progress by 1
	Increment()
	
	// IncrementBy increases the current progress by the specified amount
	IncrementBy(amount int)
	
	// SetCurrent sets the current progress to a specific value
	SetCurrent(current int)
	
	// SetMessage sets the current operation message
	SetMessage(message string)
	
	// Finish marks the progress as complete and performs any cleanup
	Finish()
	
	// SetError marks the progress as errored and displays error state
	SetError(err error)
	
	// IsComplete returns true if progress is complete
	IsComplete() bool
	
	// Current returns the current progress count
	Current() int
	
	// Total returns the total progress count
	Total() int
}

// NoOpReporter is a progress reporter that does nothing, useful for when progress is disabled.
type NoOpReporter struct{}

func NewNoOpReporter() *NoOpReporter {
	return &NoOpReporter{}
}

func (n *NoOpReporter) SetTotal(total int)         {}
func (n *NoOpReporter) Increment()                {}
func (n *NoOpReporter) IncrementBy(amount int)    {}
func (n *NoOpReporter) SetCurrent(current int)    {}
func (n *NoOpReporter) SetMessage(message string) {}
func (n *NoOpReporter) Finish()                   {}
func (n *NoOpReporter) SetError(err error)        {}
func (n *NoOpReporter) IsComplete() bool          { return false }
func (n *NoOpReporter) Current() int              { return 0 }
func (n *NoOpReporter) Total() int                { return 0 }

// ProgressState represents the current state of progress tracking.
type ProgressState struct {
	current       int
	total         int
	actualCurrent int // Track actual current before capping, for percentage calculation
	message       string
	writer        io.Writer
}

// NewProgressState creates a new progress state.
func NewProgressState(writer io.Writer) *ProgressState {
	return &ProgressState{
		writer: writer,
	}
}

// SetTotal sets the total number of items to be processed.
func (p *ProgressState) SetTotal(total int) {
	if total < 0 {
		total = 0
	}
	p.total = total
}

// SetCurrent sets the current progress.
func (p *ProgressState) SetCurrent(current int) {
	if current < 0 {
		current = 0
	}
	p.actualCurrent = current
	if current > p.total && p.total > 0 {
		current = p.total
	}
	p.current = current
}

// Increment increases the progress by 1.
func (p *ProgressState) Increment() {
	p.IncrementBy(1)
}

// IncrementBy increases the progress by the specified amount.
func (p *ProgressState) IncrementBy(amount int) {
	if amount < 0 {
		return
	}
	newCurrent := p.actualCurrent + amount
	p.SetCurrent(newCurrent)
}

// SetMessage sets the current operation message.
func (p *ProgressState) SetMessage(message string) {
	p.message = message
}

// Current returns the current progress.
func (p *ProgressState) Current() int {
	return p.current
}

// Total returns the total progress.
func (p *ProgressState) Total() int {
	return p.total
}

// Message returns the current message.
func (p *ProgressState) Message() string {
	return p.message
}

// IsComplete returns true if progress is complete.
func (p *ProgressState) IsComplete() bool {
	return p.total > 0 && p.current >= p.total
}

// Percentage returns the completion percentage (0-100).
func (p *ProgressState) Percentage() int {
	if p.total == 0 {
		return 0
	}
	return int((float64(p.actualCurrent) / float64(p.total)) * 100)
}

// String returns a string representation of the progress.
func (p *ProgressState) String() string {
	if p.total == 0 {
		return fmt.Sprintf("%d items processed", p.current)
	}
	return fmt.Sprintf("%d/%d (%d%%)", p.current, p.total, p.Percentage())
}