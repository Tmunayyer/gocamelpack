package files

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/Tmunayyer/gocamelpack/progress"
)

// mockProgressReporter for testing progress integration
type mockProgressReporter struct {
	total       int
	current     int
	messages    []string
	finished    bool
	errored     bool
	increments  int
}

func newMockProgressReporter() *mockProgressReporter {
	return &mockProgressReporter{
		messages: make([]string, 0),
	}
}

func (m *mockProgressReporter) SetTotal(total int) {
	m.total = total
}

func (m *mockProgressReporter) Increment() {
	m.current++
	m.increments++
}

func (m *mockProgressReporter) IncrementBy(amount int) {
	m.current += amount
	m.increments += amount
}

func (m *mockProgressReporter) SetCurrent(current int) {
	m.current = current
}

func (m *mockProgressReporter) SetMessage(message string) {
	m.messages = append(m.messages, message)
}

func (m *mockProgressReporter) Finish() {
	m.finished = true
}

func (m *mockProgressReporter) SetError(err error) {
	// For mock purposes, mark as errored but not finished
	m.errored = true
}

func (m *mockProgressReporter) IsComplete() bool {
	return m.total > 0 && m.current >= m.total
}

func (m *mockProgressReporter) Current() int {
	return m.current
}

func (m *mockProgressReporter) Total() int {
	return m.total
}

// mockFilesService for testing transaction progress
type mockFilesService struct {
	files         map[string]bool
	copyCallCount int
	copyError     error
	failOnCopy    int // Fail on nth copy (1-indexed, 0 = never fail)
}

func newMockFilesService() *mockFilesService {
	return &mockFilesService{
		files: make(map[string]bool),
	}
}

func (m *mockFilesService) addFile(path string) {
	m.files[path] = true
}

func (m *mockFilesService) setFailOnCopy(n int) {
	m.failOnCopy = n
}

func (m *mockFilesService) Copy(src, dst string) error {
	m.copyCallCount++
	if m.failOnCopy > 0 && m.copyCallCount == m.failOnCopy {
		return errors.New("mock copy error")
	}
	return m.copyError
}

func (m *mockFilesService) ValidateCopyArgs(src, dst string) error {
	if !m.files[src] {
		return errors.New("source file does not exist")
	}
	return nil
}

func (m *mockFilesService) IsFile(path string) bool {
	return m.files[path]
}

func (m *mockFilesService) EnsureDir(path string, perm os.FileMode) error {
	return nil
}

func (m *mockFilesService) IsDirectory(path string) bool {
	return false
}

func (m *mockFilesService) GetFileTags(paths []string) []FileMetadata {
	return nil
}

func (m *mockFilesService) ReadDirectory(dirPath string) ([]string, error) {
	return nil, nil
}

func (m *mockFilesService) DestinationFromMetadata(tags FileMetadata, baseDir string) (string, error) {
	return "", nil
}

func (m *mockFilesService) Close() {
	// No-op for mock
}

func (m *mockFilesService) NewTransaction(overwrite bool) Transaction {
	return NewTransaction(m, overwrite)
}

func TestFileTransaction_ExecuteWithProgress_Success(t *testing.T) {
	mockFS := newMockFilesService()
	mockFS.addFile("/src/file1.txt")
	mockFS.addFile("/src/file2.txt")
	mockFS.addFile("/src/file3.txt")

	tx := NewTransaction(mockFS, false)
	
	// Add operations
	err := tx.AddCopy("/src/file1.txt", "/dst/file1.txt")
	if err != nil {
		t.Fatalf("AddCopy failed: %v", err)
	}
	
	err = tx.AddCopy("/src/file2.txt", "/dst/file2.txt")
	if err != nil {
		t.Fatalf("AddCopy failed: %v", err)
	}
	
	err = tx.AddCopy("/src/file3.txt", "/dst/file3.txt")
	if err != nil {
		t.Fatalf("AddCopy failed: %v", err)
	}
	
	// Execute with progress
	mockReporter := newMockProgressReporter()
	err = tx.ExecuteWithProgress(mockReporter)
	if err != nil {
		t.Fatalf("ExecuteWithProgress failed: %v", err)
	}
	
	// Verify progress tracking
	if mockReporter.total != 3 {
		t.Errorf("Expected total=3, got %d", mockReporter.total)
	}
	
	if mockReporter.current != 3 {
		t.Errorf("Expected current=3, got %d", mockReporter.current)
	}
	
	if !mockReporter.finished {
		t.Error("Expected reporter to be finished")
	}
	
	// Verify messages were set (one per operation)
	if len(mockReporter.messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(mockReporter.messages))
	}
	
	// Verify message content
	expectedMessages := []string{
		"copy /src/file1.txt",
		"copy /src/file2.txt", 
		"copy /src/file3.txt",
	}
	
	for i, expected := range expectedMessages {
		if i >= len(mockReporter.messages) {
			t.Errorf("Missing message %d", i)
			continue
		}
		if mockReporter.messages[i] != expected {
			t.Errorf("Message %d: expected %q, got %q", i, expected, mockReporter.messages[i])
		}
	}
}

func TestFileTransaction_ExecuteWithProgress_Failure(t *testing.T) {
	mockFS := newMockFilesService()
	mockFS.addFile("/src/file1.txt")
	mockFS.addFile("/src/file2.txt")
	mockFS.setFailOnCopy(2) // Fail on second copy

	tx := NewTransaction(mockFS, false)
	
	// Add operations
	err := tx.AddCopy("/src/file1.txt", "/dst/file1.txt")
	if err != nil {
		t.Fatalf("AddCopy failed: %v", err)
	}
	
	err = tx.AddCopy("/src/file2.txt", "/dst/file2.txt")
	if err != nil {
		t.Fatalf("AddCopy failed: %v", err)
	}
	
	// Execute with progress - should fail
	mockReporter := newMockProgressReporter()
	err = tx.ExecuteWithProgress(mockReporter)
	if err == nil {
		t.Fatal("Expected ExecuteWithProgress to fail")
	}
	
	// Verify progress was updated before failure
	if mockReporter.total != 2 {
		t.Errorf("Expected total=2, got %d", mockReporter.total)
	}
	
	// Should not finish on error
	if mockReporter.finished {
		t.Error("Expected reporter NOT to be finished on error")
	}
	
	// Should have at least one message (for the first operation)
	if len(mockReporter.messages) < 1 {
		t.Error("Expected at least one progress message before failure")
	}
}

func TestFileTransaction_Execute_UsesNoOpReporter(t *testing.T) {
	mockFS := newMockFilesService()
	mockFS.addFile("/src/file1.txt")

	tx := NewTransaction(mockFS, false)
	
	err := tx.AddCopy("/src/file1.txt", "/dst/file1.txt")
	if err != nil {
		t.Fatalf("AddCopy failed: %v", err)
	}
	
	// Execute without progress should work (uses NoOpReporter internally)
	err = tx.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	
	// Verify the operation was completed
	completed := tx.Completed()
	if len(completed) != 1 {
		t.Errorf("Expected 1 completed operation, got %d", len(completed))
	}
}

func TestFileTransaction_ExecuteWithProgress_EmptyTransaction(t *testing.T) {
	mockFS := newMockFilesService()
	tx := NewTransaction(mockFS, false)
	
	mockReporter := newMockProgressReporter()
	err := tx.ExecuteWithProgress(mockReporter)
	if err != nil {
		t.Fatalf("ExecuteWithProgress failed on empty transaction: %v", err)
	}
	
	// Should set total to 0 and finish immediately
	if mockReporter.total != 0 {
		t.Errorf("Expected total=0, got %d", mockReporter.total)
	}
	
	if mockReporter.current != 0 {
		t.Errorf("Expected current=0, got %d", mockReporter.current)
	}
	
	if !mockReporter.finished {
		t.Error("Expected reporter to be finished even for empty transaction")
	}
}

func TestFileTransaction_ExecuteWithProgress_MoveOperations(t *testing.T) {
	// Skip this test since MoveOperation.Execute uses os.Rename which requires real files
	// This test would be better as an integration test with a real filesystem
	t.Skip("Skipping move operation test - requires real filesystem for os.Rename")
}

func TestFileTransaction_ExecuteWithProgress_IntegrationWithProgressBar(t *testing.T) {
	mockFS := newMockFilesService()
	mockFS.addFile("/src/file1.txt")
	mockFS.addFile("/src/file2.txt")

	tx := NewTransaction(mockFS, false)
	
	err := tx.AddCopy("/src/file1.txt", "/dst/file1.txt")
	if err != nil {
		t.Fatalf("AddCopy failed: %v", err)
	}
	
	err = tx.AddCopy("/src/file2.txt", "/dst/file2.txt")
	if err != nil {
		t.Fatalf("AddCopy failed: %v", err)
	}
	
	// Test with actual ProgressBar
	buf := &bytes.Buffer{}
	progressBar := progress.NewProgressBar(buf, 20)
	progressBar.SetBarChar('=')
	progressBar.SetEmptyChar('-')
	
	err = tx.ExecuteWithProgress(progressBar)
	if err != nil {
		t.Fatalf("ExecuteWithProgress with ProgressBar failed: %v", err)
	}
	
	// Verify the progress bar output contains expected elements
	output := buf.String()
	
	// Should contain progress bar elements
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("Expected progress bar brackets in output")
	}
	
	// Should contain percentage
	if !strings.Contains(output, "%") {
		t.Error("Expected percentage in progress bar output")
	}
	
	// Should contain checkmark at the end (from Finish())
	if !strings.Contains(output, "✓") {
		t.Error("Expected checkmark in final progress bar output")
	}
}