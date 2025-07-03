package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Tmunayyer/gocamelpack/progress"
	"github.com/Tmunayyer/gocamelpack/testutil"
)

func TestCollectSourcesWithProgress_SingleFile(t *testing.T) {
	tempDir := testutil.TempDir(t)
	testFile := filepath.Join(tempDir, "test.jpg")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	filesService := createTestFilesService(nil)
	
	// Test with progress reporter
	buf := &bytes.Buffer{}
	reporter := progress.NewProgressBar(buf, 20)
	reporter.SetBarChar('=')
	reporter.SetEmptyChar('-')

	sources, err := collectSourcesWithProgress(filesService, testFile, reporter)
	if err != nil {
		t.Fatalf("collectSourcesWithProgress failed: %v", err)
	}

	// Verify correct source collected
	if len(sources) != 1 {
		t.Fatalf("Expected 1 source, got %d", len(sources))
	}

	if !strings.Contains(sources[0], "test.jpg") {
		t.Errorf("Expected source to contain test.jpg, got %s", sources[0])
	}

	// Verify progress output
	output := buf.String()
	if !strings.Contains(output, "Collecting single file") {
		t.Error("Expected progress message for single file collection")
	}

	if !strings.Contains(output, "✓") {
		t.Error("Expected completion checkmark in progress output")
	}
}

func TestCollectSourcesWithProgress_Directory(t *testing.T) {
	tempDir := testutil.TempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files
	testFiles := []string{"test1.jpg", "test2.jpg", "test3.jpg"}
	for _, filename := range testFiles {
		testFile := filepath.Join(srcDir, filename)
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	filesService := createTestFilesService(nil)
	
	// Test with progress reporter
	buf := &bytes.Buffer{}
	reporter := progress.NewProgressBar(buf, 20)
	reporter.SetBarChar('=')
	reporter.SetEmptyChar('-')

	sources, err := collectSourcesWithProgress(filesService, srcDir, reporter)
	if err != nil {
		t.Fatalf("collectSourcesWithProgress failed: %v", err)
	}

	// Verify correct sources collected
	if len(sources) != 3 {
		t.Fatalf("Expected 3 sources, got %d", len(sources))
	}

	// Verify progress output contains expected messages
	output := buf.String()
	if !strings.Contains(output, "Reading directory") {
		t.Error("Expected 'Reading directory' message in progress output")
	}

	if !strings.Contains(output, "Collecting files from directory") {
		t.Error("Expected 'Collecting files from directory' message in progress output")
	}

	if !strings.Contains(output, "✓") {
		t.Error("Expected completion checkmark in progress output")
	}

	// Should contain progress bar elements
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("Expected progress bar brackets in output")
	}
}

func TestCollectSourcesWithProgress_EmptyDirectory(t *testing.T) {
	tempDir := testutil.TempDir(t)
	srcDir := filepath.Join(tempDir, "empty")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	filesService := createTestFilesService(nil)
	
	// Test with progress reporter
	buf := &bytes.Buffer{}
	reporter := progress.NewProgressBar(buf, 20)

	sources, err := collectSourcesWithProgress(filesService, srcDir, reporter)
	if err != nil {
		t.Fatalf("collectSourcesWithProgress failed: %v", err)
	}

	// Verify no sources collected
	if len(sources) != 0 {
		t.Fatalf("Expected 0 sources, got %d", len(sources))
	}

	// Verify progress output shows completion even for empty directory
	output := buf.String()
	if !strings.Contains(output, "Reading directory") {
		t.Error("Expected 'Reading directory' message in progress output")
	}

	if !strings.Contains(output, "✓") {
		t.Error("Expected completion checkmark even for empty directory")
	}
}

func TestCollectSources_BackwardCompatibility(t *testing.T) {
	tempDir := testutil.TempDir(t)
	testFile := filepath.Join(tempDir, "test.jpg")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	filesService := createTestFilesService(nil)
	
	// Test that original function still works (uses NoOpReporter internally)
	sources, err := collectSources(filesService, testFile)
	if err != nil {
		t.Fatalf("collectSources failed: %v", err)
	}

	// Verify correct source collected
	if len(sources) != 1 {
		t.Fatalf("Expected 1 source, got %d", len(sources))
	}

	if !strings.Contains(sources[0], "test.jpg") {
		t.Errorf("Expected source to contain test.jpg, got %s", sources[0])
	}
}

func TestCollectSourcesWithProgress_NoOpReporter(t *testing.T) {
	tempDir := testutil.TempDir(t)
	testFile := filepath.Join(tempDir, "test.jpg")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	filesService := createTestFilesService(nil)
	
	// Test with NoOpReporter - should work without issues
	reporter := progress.NewNoOpReporter()
	sources, err := collectSourcesWithProgress(filesService, testFile, reporter)
	if err != nil {
		t.Fatalf("collectSourcesWithProgress with NoOpReporter failed: %v", err)
	}

	// Verify correct source collected
	if len(sources) != 1 {
		t.Fatalf("Expected 1 source, got %d", len(sources))
	}

	// NoOpReporter should always return false for IsComplete
	if reporter.IsComplete() {
		t.Error("NoOpReporter should never report as complete")
	}
}

func TestCollectSourcesWithProgress_InvalidPath(t *testing.T) {
	filesService := createTestFilesService(nil)
	
	buf := &bytes.Buffer{}
	reporter := progress.NewProgressBar(buf, 20)

	// Test with non-existent path
	_, err := collectSourcesWithProgress(filesService, "/nonexistent/path", reporter)
	if err == nil {
		t.Error("Expected collectSourcesWithProgress to fail with invalid path")
	}

	// Error should occur before any progress reporting
	output := buf.String()
	// Should not have completed successfully
	if strings.Contains(output, "✓") {
		t.Error("Should not show completion checkmark for failed operation")
	}
}