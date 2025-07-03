package progress

import (
	"fmt"
	"io"
	"strings"
)

// ProgressBar implements a visual ASCII progress bar.
type ProgressBar struct {
	*ProgressState
	width     int
	barChar   rune
	emptyChar rune
	showMsg   bool
	finished  bool
	errored   bool
}

// NewProgressBar creates a new progress bar with the specified width and output writer.
func NewProgressBar(writer io.Writer, width int) *ProgressBar {
	if width <= 0 {
		width = 40 // Default width
	}
	
	return &ProgressBar{
		ProgressState: NewProgressState(writer),
		width:         width,
		barChar:       '█',
		emptyChar:     '░',
		showMsg:       true,
	}
}

// NewSimpleProgressBar creates a basic progress bar with default settings.
func NewSimpleProgressBar(writer io.Writer) *ProgressBar {
	return NewProgressBar(writer, 40)
}

// SetBarChar sets the character used for the filled portion of the bar.
func (pb *ProgressBar) SetBarChar(char rune) {
	pb.barChar = char
}

// SetEmptyChar sets the character used for the empty portion of the bar.
func (pb *ProgressBar) SetEmptyChar(char rune) {
	pb.emptyChar = char
}

// SetShowMessage controls whether the current message is displayed.
func (pb *ProgressBar) SetShowMessage(show bool) {
	pb.showMsg = show
}

// SetWidth sets the width of the progress bar in characters.
func (pb *ProgressBar) SetWidth(width int) {
	if width > 0 {
		pb.width = width
	}
}

// Render returns the current progress bar as a string without printing it.
func (pb *ProgressBar) Render() string {
	if pb.finished || pb.errored {
		return "" // Don't render after finish or error
	}
	
	var result strings.Builder
	
	// Calculate filled portion
	var filledWidth int
	if pb.total > 0 {
		filledWidth = int(float64(pb.current) / float64(pb.total) * float64(pb.width))
	}
	
	// Ensure filled width doesn't exceed bar width
	if filledWidth > pb.width {
		filledWidth = pb.width
	}
	
	// Build the bar
	result.WriteRune('[')
	
	// Filled portion
	for i := 0; i < filledWidth; i++ {
		result.WriteRune(pb.barChar)
	}
	
	// Empty portion
	for i := filledWidth; i < pb.width; i++ {
		result.WriteRune(pb.emptyChar)
	}
	
	result.WriteRune(']')
	
	// Add percentage and counts
	result.WriteString(fmt.Sprintf(" %s", pb.String()))
	
	// Add message if enabled and present
	if pb.showMsg && pb.message != "" {
		result.WriteString(fmt.Sprintf(" - %s", pb.message))
	}
	
	return result.String()
}

// Display renders and prints the progress bar to the configured writer.
func (pb *ProgressBar) Display() {
	if pb.finished || pb.errored {
		return
	}
	
	rendered := pb.Render()
	if rendered != "" {
		fmt.Fprint(pb.writer, "\r"+rendered)
	}
}

// Update increments progress and displays the updated bar.
func (pb *ProgressBar) Update() {
	pb.Display()
}

// Finish marks the progress as complete and displays a final message.
func (pb *ProgressBar) Finish() {
	if pb.finished {
		return
	}
	
	pb.finished = true
	
	// Show final state
	var result strings.Builder
	
	// Build completed bar
	result.WriteRune('[')
	for i := 0; i < pb.width; i++ {
		result.WriteRune(pb.barChar)
	}
	result.WriteRune(']')
	
	// Add final stats
	result.WriteString(fmt.Sprintf(" %s", pb.String()))
	
	if pb.showMsg && pb.message != "" {
		result.WriteString(fmt.Sprintf(" - %s", pb.message))
	}
	
	result.WriteString(" ✓\n") // Checkmark and newline to finish
	
	fmt.Fprint(pb.writer, "\r"+result.String())
}

// Increment increases progress by 1 and updates the display.
func (pb *ProgressBar) Increment() {
	if pb.errored || pb.finished {
		return
	}
	pb.ProgressState.Increment()
	pb.Update()
}

// IncrementBy increases progress by the specified amount and updates the display.
func (pb *ProgressBar) IncrementBy(amount int) {
	if pb.errored || pb.finished {
		return
	}
	pb.ProgressState.IncrementBy(amount)
	pb.Update()
}

// SetCurrent sets the current progress and updates the display.
func (pb *ProgressBar) SetCurrent(current int) {
	if pb.errored || pb.finished {
		return
	}
	pb.ProgressState.SetCurrent(current)
	pb.Update()
}

// SetTotal sets the total and updates the display.
func (pb *ProgressBar) SetTotal(total int) {
	if pb.errored || pb.finished {
		return
	}
	pb.ProgressState.SetTotal(total)
	pb.Update()
}

// SetMessage sets the message and updates the display.
func (pb *ProgressBar) SetMessage(message string) {
	if pb.errored || pb.finished {
		return
	}
	pb.ProgressState.SetMessage(message)
	pb.Update()
}

// SetError marks the progress bar as errored and displays an error state.
func (pb *ProgressBar) SetError(err error) {
	if pb.finished {
		return
	}
	
	pb.errored = true
	pb.finished = true
	
	// Show error state
	var result strings.Builder
	
	// Build error bar - show current progress with error indicator
	result.WriteRune('[')
	
	var filledWidth int
	if pb.total > 0 {
		filledWidth = int(float64(pb.current) / float64(pb.total) * float64(pb.width))
	}
	if filledWidth > pb.width {
		filledWidth = pb.width
	}
	
	// Filled portion
	for i := 0; i < filledWidth; i++ {
		result.WriteRune(pb.barChar)
	}
	
	// Empty portion
	for i := filledWidth; i < pb.width; i++ {
		result.WriteRune(pb.emptyChar)
	}
	
	result.WriteRune(']')
	
	// Add current stats
	result.WriteString(fmt.Sprintf(" %s", pb.String()))
	
	if pb.showMsg && pb.message != "" {
		result.WriteString(fmt.Sprintf(" - %s", pb.message))
	}
	
	result.WriteString(" ✗") // Error mark
	if err != nil {
		result.WriteString(fmt.Sprintf(" - Error: %s", err.Error()))
	}
	result.WriteString("\n")
	
	fmt.Fprint(pb.writer, "\r"+result.String())
}

// IsErrored returns true if the progress bar is in an error state.
func (pb *ProgressBar) IsErrored() bool {
	return pb.errored
}