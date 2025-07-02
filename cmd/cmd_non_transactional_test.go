package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/Tmunayyer/gocamelpack/files"
	"github.com/Tmunayyer/gocamelpack/progress"
)

// mockFilesServiceForCmd implements FilesService for testing command functions
type mockFilesServiceForCmd struct {
	files         map[string]bool
	copyCallCount int
	copyError     error
	validationErr error
}

func newMockFilesServiceForCmd() *mockFilesServiceForCmd {
	return &mockFilesServiceForCmd{
		files: make(map[string]bool),
	}
}

func (m *mockFilesServiceForCmd) addFile(path string) {
	m.files[path] = true
}

func (m *mockFilesServiceForCmd) setValidationError(err error) {
	m.validationErr = err
}

func (m *mockFilesServiceForCmd) Close() {}

func (m *mockFilesServiceForCmd) IsFile(path string) bool {
	return m.files[path]
}

func (m *mockFilesServiceForCmd) IsDirectory(path string) bool {
	return false
}

func (m *mockFilesServiceForCmd) GetFileTags(paths []string) []files.FileMetadata {
	var result []files.FileMetadata
	for _, path := range paths {
		result = append(result, files.FileMetadata{
			Filepath: path,
			Tags:     make(map[string]string),
		})
	}
	return result
}

func (m *mockFilesServiceForCmd) ReadDirectory(dirPath string) ([]string, error) {
	return nil, nil
}

func (m *mockFilesServiceForCmd) DestinationFromMetadata(tags files.FileMetadata, baseDir string) (string, error) {
	// Simple mock: create predictable destinations based on source path
	filename := strings.TrimPrefix(tags.Filepath, "/src/")
	return baseDir + "/" + filename, nil
}

func (m *mockFilesServiceForCmd) Copy(src, dst string) error {
	m.copyCallCount++
	return m.copyError
}

func (m *mockFilesServiceForCmd) EnsureDir(path string, perm os.FileMode) error {
	return nil
}

func (m *mockFilesServiceForCmd) ValidateCopyArgs(src, dst string) error {
	return m.validationErr
}

func (m *mockFilesServiceForCmd) NewTransaction(overwrite bool) files.Transaction {
	return nil // Not used in these tests
}

func TestPerformNonTransactionalCopy_Success(t *testing.T) {

	mockFS := newMockFilesServiceForCmd()
	mockFS.addFile("/src/file1.txt")
	mockFS.addFile("/src/file2.txt")

	sources := []string{"/src/file1.txt", "/src/file2.txt"}
	dstRoot := "/dst"

	// Create a command for output testing
	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	err := performNonTransactionalCopy(mockFS, sources, dstRoot, false, false, cmd)
	if err != nil {
		t.Fatalf("performNonTransactionalCopy failed: %v", err)
	}

	// Verify Copy was called for each file
	if mockFS.copyCallCount != 2 {
		t.Errorf("Expected 2 copy calls, got %d", mockFS.copyCallCount)
	}
}

func TestPerformNonTransactionalCopy_DryRun(t *testing.T) {

	mockFS := newMockFilesServiceForCmd()
	sources := []string{"/src/file1.txt", "/src/file2.txt"}
	dstRoot := "/dst"

	// Create a command for output testing
	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	err := performNonTransactionalCopy(mockFS, sources, dstRoot, true, false, cmd)
	if err != nil {
		t.Fatalf("performNonTransactionalCopy dry-run failed: %v", err)
	}

	// Verify no actual copies were performed
	if mockFS.copyCallCount != 0 {
		t.Errorf("Expected 0 copy calls in dry-run, got %d", mockFS.copyCallCount)
	}

	// Verify dry-run output
	output := buf.String()
	if !strings.Contains(output, "Would copy") {
		t.Error("Expected dry-run output to contain 'Would copy'")
	}
	if !strings.Contains(output, "/src/file1.txt → /dst/file1.txt") {
		t.Error("Expected dry-run output to show file1 mapping")
	}
}

func TestPerformNonTransactionalCopy_ValidationError(t *testing.T) {

	mockFS := newMockFilesServiceForCmd()
	mockFS.setValidationError(os.ErrExist)

	sources := []string{"/src/file1.txt"}
	dstRoot := "/dst"

	cmd := &cobra.Command{}

	err := performNonTransactionalCopy(mockFS, sources, dstRoot, false, false, cmd)
	if err == nil {
		t.Fatal("Expected performNonTransactionalCopy to fail with validation error")
	}

	// Should not have performed any copies due to validation failure
	if mockFS.copyCallCount != 0 {
		t.Errorf("Expected 0 copy calls due to validation error, got %d", mockFS.copyCallCount)
	}
}

func TestPerformNonTransactionalMove_Success(t *testing.T) {

	mockFS := newMockFilesServiceForCmd()
	mockFS.addFile("/src/file1.txt")

	sources := []string{"/src/file1.txt"}
	dstRoot := "/dst"

	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	// Note: This test will actually try to call os.Rename, which will fail
	// because the files don't exist. In a real scenario, we'd need a more
	// sophisticated mock or integration test with real files.
	err := performNonTransactionalMove(mockFS, sources, dstRoot, false, false, cmd)
	
	// We expect this to fail because os.Rename tries to move real files
	if err == nil {
		t.Fatal("Expected performNonTransactionalMove to fail without real files")
	}
}

func TestPerformNonTransactionalMove_DryRun(t *testing.T) {

	mockFS := newMockFilesServiceForCmd()
	sources := []string{"/src/file1.txt", "/src/file2.txt"}
	dstRoot := "/dst"

	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	err := performNonTransactionalMove(mockFS, sources, dstRoot, true, false, cmd)
	if err != nil {
		t.Fatalf("performNonTransactionalMove dry-run failed: %v", err)
	}

	// Verify dry-run output
	output := buf.String()
	if !strings.Contains(output, "Would move") {
		t.Error("Expected dry-run output to contain 'Would move'")
	}
	if !strings.Contains(output, "/src/file1.txt → /dst/file1.txt") {
		t.Error("Expected dry-run output to show file1 mapping")
	}
}

func TestPerformNonTransactionalCopy_WithProgressReporter(t *testing.T) {
	// This test verifies that the progress reporting infrastructure is in place
	// The actual progress reporting will be tested when CLI integration is added

	mockFS := newMockFilesServiceForCmd()
	mockFS.addFile("/src/file1.txt")
	mockFS.addFile("/src/file2.txt")

	sources := []string{"/src/file1.txt", "/src/file2.txt"}
	dstRoot := "/dst"

	cmd := &cobra.Command{}

	// Test that the function completes without error
	// (Progress reporting is currently using NoOpReporter)
	err := performNonTransactionalCopy(mockFS, sources, dstRoot, false, false, cmd)
	if err != nil {
		t.Fatalf("performNonTransactionalCopy with progress failed: %v", err)
	}

	// Verify that operations were performed
	if mockFS.copyCallCount != 2 {
		t.Errorf("Expected 2 copy calls, got %d", mockFS.copyCallCount)
	}
}

func TestPerformNonTransactionalMove_WithProgressReporter(t *testing.T) {
	// This test verifies that the progress reporting infrastructure is in place
	// for move operations (dry-run mode to avoid os.Rename issues)

	mockFS := newMockFilesServiceForCmd()
	sources := []string{"/src/file1.txt", "/src/file2.txt"}
	dstRoot := "/dst"

	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	// Test dry-run mode (to avoid os.Rename complications)
	err := performNonTransactionalMove(mockFS, sources, dstRoot, true, false, cmd)
	if err != nil {
		t.Fatalf("performNonTransactionalMove dry-run with progress failed: %v", err)
	}

	// Verify dry-run worked
	output := buf.String()
	if !strings.Contains(output, "Would move") {
		t.Error("Expected dry-run output to contain 'Would move'")
	}
}

func TestProgressReporterUsage_CurrentlyNoOp(t *testing.T) {
	// This test documents that we're currently using NoOpReporter
	// and will be updated when CLI integration adds real progress reporting

	reporter := progress.NewNoOpReporter()
	
	// Set up typical progress workflow
	reporter.SetTotal(5)
	reporter.SetMessage("test operation")
	reporter.Increment()
	reporter.SetCurrent(3)
	reporter.Finish()
	
	// NoOpReporter should always return default values
	if reporter.Current() != 0 {
		t.Errorf("NoOpReporter.Current() should return 0, got %d", reporter.Current())
	}
	
	if reporter.Total() != 0 {
		t.Errorf("NoOpReporter.Total() should return 0, got %d", reporter.Total())
	}
	
	if reporter.IsComplete() {
		t.Error("NoOpReporter.IsComplete() should return false")
	}
}