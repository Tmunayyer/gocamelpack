package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Tmunayyer/gocamelpack/deps"
	"github.com/Tmunayyer/gocamelpack/testutil"
)

func TestCopyCmd_WithProgress_ShowsCollectionProgress(t *testing.T) {
	tempDir := testutil.TempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create multiple test files to see collection progress
	testFiles := []string{"test1.jpg", "test2.jpg", "test3.jpg", "test4.jpg", "test5.jpg"}
	for _, filename := range testFiles {
		testFile := filepath.Join(srcDir, filename)
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	filesService := createTestFilesService(nil)
	dep := &deps.AppDeps{Files: filesService}
	cmd := createCopyCmd(dep)
	cmd.SetArgs([]string{"--progress", "--overwrite", srcDir, dstDir})

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("copy command with progress failed: %v", err)
	}

	// Verify collection progress appears in stderr
	stderrOutput := stderr.String()
	
	// Should contain collection phase messages
	if !strings.Contains(stderrOutput, "Reading directory") {
		t.Error("Expected 'Reading directory' message in stderr output")
	}
	
	if !strings.Contains(stderrOutput, "Collecting files") {
		t.Error("Expected 'Collecting files' message in stderr output")
	}
	
	// Should contain copy execution messages
	if !strings.Contains(stderrOutput, "copy") {
		t.Error("Expected copy operation messages in stderr output")
	}
	
	// Should contain multiple completion checkmarks (collection + execution)
	checkmarkCount := strings.Count(stderrOutput, "✓")
	if checkmarkCount < 2 {
		t.Errorf("Expected at least 2 completion checkmarks (collection + execution), got %d", checkmarkCount)
	}

	// Verify completion message in stdout
	stdoutOutput := out.String()
	if !strings.Contains(stdoutOutput, "Copied") {
		t.Errorf("Expected completion message in stdout, got: %q", stdoutOutput)
	}
}

func TestCopyCmd_WithProgress_AtomicShowsPlanningProgress(t *testing.T) {
	tempDir := testutil.TempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
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
	dep := &deps.AppDeps{Files: filesService}
	cmd := createCopyCmd(dep)
	cmd.SetArgs([]string{"--atomic", "--progress", "--overwrite", srcDir, dstDir})

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("atomic copy command with progress failed: %v", err)
	}

	// Verify all progress phases appear
	stderrOutput := stderr.String()
	
	// Collection phase
	if !strings.Contains(stderrOutput, "Reading directory") {
		t.Error("Expected collection progress in stderr output")
	}
	
	// Planning phase
	if !strings.Contains(stderrOutput, "Planning copy for") {
		t.Error("Expected planning progress with file details in stderr output")
	}
	
	// Execution phase
	if !strings.Contains(stderrOutput, "copy") {
		t.Error("Expected execution progress in stderr output")
	}
	
	// Should have multiple progress bars completed
	checkmarkCount := strings.Count(stderrOutput, "✓")
	if checkmarkCount < 3 {
		t.Errorf("Expected at least 3 completion checkmarks (collection + planning + execution), got %d", checkmarkCount)
	}

	// Verify atomic completion message
	stdoutOutput := out.String()
	if !strings.Contains(stdoutOutput, "Atomically copied") {
		t.Error("Expected atomic completion message in stdout")
	}
}

func TestMoveCmd_WithProgress_ShowsPlanningProgress(t *testing.T) {
	tempDir := testutil.TempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files
	testFiles := []string{"test1.jpg", "test2.jpg"}
	for _, filename := range testFiles {
		testFile := filepath.Join(srcDir, filename)
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	filesService := createTestFilesService(nil)
	dep := &deps.AppDeps{Files: filesService}
	cmd := createMoveCmd(dep)
	cmd.SetArgs([]string{"--atomic", "--progress", "--overwrite", srcDir, dstDir})

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("atomic move command with progress failed: %v", err)
	}

	// Verify planning progress for move operations
	stderrOutput := stderr.String()
	
	if !strings.Contains(stderrOutput, "Planning move for") {
		t.Error("Expected planning progress for move operations in stderr output")
	}
	
	if !strings.Contains(stderrOutput, "move") {
		t.Error("Expected move execution progress in stderr output")
	}

	// Verify atomic completion message
	stdoutOutput := out.String()
	if !strings.Contains(stdoutOutput, "Atomically moved") {
		t.Error("Expected atomic move completion message in stdout")
	}
}

func TestCopyCmd_WithoutProgress_NoCollectionProgress(t *testing.T) {
	tempDir := testutil.TempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files
	testFiles := []string{"test1.jpg", "test2.jpg"}
	for _, filename := range testFiles {
		testFile := filepath.Join(srcDir, filename)
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	filesService := createTestFilesService(nil)
	dep := &deps.AppDeps{Files: filesService}
	cmd := createCopyCmd(dep)
	cmd.SetArgs([]string{"--overwrite", srcDir, dstDir}) // No --progress flag

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("copy command without progress failed: %v", err)
	}

	// Verify NO progress output in stderr
	stderrOutput := stderr.String()
	if strings.Contains(stderrOutput, "Reading directory") || 
	   strings.Contains(stderrOutput, "Planning") ||
	   strings.Contains(stderrOutput, "✓") {
		t.Errorf("Expected no progress output without --progress flag, got: %q", stderrOutput)
	}

	// Should still have completion message in stdout
	stdoutOutput := out.String()
	if !strings.Contains(stdoutOutput, "Copied") {
		t.Error("Expected completion message in stdout")
	}
}

func TestCopyCmd_SingleFile_MinimalCollectionProgress(t *testing.T) {
	tempDir := testutil.TempDir(t)
	testFile := filepath.Join(tempDir, "test.jpg")
	dstDir := filepath.Join(tempDir, "dst")

	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatal(err)
	}

	filesService := createTestFilesService(nil)
	dep := &deps.AppDeps{Files: filesService}
	cmd := createCopyCmd(dep)
	cmd.SetArgs([]string{"--progress", "--overwrite", testFile, dstDir})

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("copy single file with progress failed: %v", err)
	}

	// Verify single file collection progress
	stderrOutput := stderr.String()
	
	if !strings.Contains(stderrOutput, "Collecting single file") {
		t.Error("Expected 'Collecting single file' message for single file operation")
	}
	
	// Should not contain directory-specific messages
	if strings.Contains(stderrOutput, "Reading directory") {
		t.Error("Should not contain directory messages for single file operation")
	}

	// Should contain execution progress
	if !strings.Contains(stderrOutput, "copy") {
		t.Error("Expected copy execution progress")
	}
}