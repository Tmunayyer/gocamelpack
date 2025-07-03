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

func TestCopyCmd_WithProgress(t *testing.T) {
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
	cmd.SetArgs([]string{"--progress", "--overwrite", srcDir, dstDir})

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("copy command with progress failed: %v", err)
	}

	// Verify progress output was written to stderr
	stderrOutput := stderr.String()
	
	// Progress bar should contain brackets and percentage
	if !strings.Contains(stderrOutput, "[") || !strings.Contains(stderrOutput, "]") {
		t.Errorf("Expected progress bar brackets in stderr output, got: %q", stderrOutput)
	}
	
	if !strings.Contains(stderrOutput, "%") {
		t.Errorf("Expected percentage in progress bar output, got: %q", stderrOutput)
	}
	
	// Should contain checkmark indicating completion
	if !strings.Contains(stderrOutput, "✓") {
		t.Errorf("Expected completion checkmark in progress bar output, got: %q", stderrOutput)
	}

	// Verify completion message in stdout
	stdoutOutput := out.String()
	if !strings.Contains(stdoutOutput, "Copied") {
		t.Errorf("Expected completion message in stdout, got: %q", stdoutOutput)
	}
}

func TestCopyCmd_WithoutProgress(t *testing.T) {
	tempDir := testutil.TempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test file
	testFile := filepath.Join(srcDir, "test.jpg")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	filesService := createTestFilesService(nil)
	dep := &deps.AppDeps{Files: filesService}
	cmd := createCopyCmd(dep)
	cmd.SetArgs([]string{"--overwrite", testFile, dstDir}) // No --progress flag

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("copy command without progress failed: %v", err)
	}

	// Verify NO progress output in stderr (should be empty or minimal)
	stderrOutput := stderr.String()
	if strings.Contains(stderrOutput, "[") || strings.Contains(stderrOutput, "✓") {
		t.Errorf("Expected no progress bar output without --progress flag, got: %q", stderrOutput)
	}

	// Verify completion message in stdout
	stdoutOutput := out.String()
	if !strings.Contains(stdoutOutput, "Copied") {
		t.Error("Expected completion message in stdout")
	}
}

func TestMoveCmd_WithProgress(t *testing.T) {
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
	cmd.SetArgs([]string{"--progress", "--overwrite", srcDir, dstDir})

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("move command with progress failed: %v", err)
	}

	// Verify progress output was written to stderr
	stderrOutput := stderr.String()
	
	// Progress bar should contain brackets and percentage
	if !strings.Contains(stderrOutput, "[") || !strings.Contains(stderrOutput, "]") {
		t.Error("Expected progress bar brackets in stderr output")
	}
	
	if !strings.Contains(stderrOutput, "%") {
		t.Error("Expected percentage in progress bar output")
	}
	
	// Should contain checkmark indicating completion
	if !strings.Contains(stderrOutput, "✓") {
		t.Error("Expected completion checkmark in progress bar output")
	}

	// Verify completion message in stdout
	stdoutOutput := out.String()
	if !strings.Contains(stdoutOutput, "Moved") && !strings.Contains(stdoutOutput, "file(s)") {
		t.Error("Expected completion message in stdout")
	}
}

func TestCopyCmd_AtomicWithProgress(t *testing.T) {
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
	cmd.SetArgs([]string{"--atomic", "--progress", "--overwrite", srcDir, dstDir})

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("atomic copy command with progress failed: %v", err)
	}

	// Verify progress output was written to stderr
	stderrOutput := stderr.String()
	
	// Progress bar should contain brackets and percentage
	if !strings.Contains(stderrOutput, "[") || !strings.Contains(stderrOutput, "]") {
		t.Error("Expected progress bar brackets in stderr output for atomic operation")
	}
	
	// Should contain checkmark indicating completion
	if !strings.Contains(stderrOutput, "✓") {
		t.Error("Expected completion checkmark in progress bar output for atomic operation")
	}

	// Verify atomic completion message in stdout
	stdoutOutput := out.String()
	if !strings.Contains(stdoutOutput, "Atomically copied") {
		t.Error("Expected atomic completion message in stdout")
	}
}

func TestMoveCmd_AtomicWithProgress(t *testing.T) {
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

	// Verify progress output was written to stderr
	stderrOutput := stderr.String()
	
	// Progress bar should contain brackets and percentage
	if !strings.Contains(stderrOutput, "[") || !strings.Contains(stderrOutput, "]") {
		t.Error("Expected progress bar brackets in stderr output for atomic operation")
	}
	
	// Should contain checkmark indicating completion
	if !strings.Contains(stderrOutput, "✓") {
		t.Error("Expected completion checkmark in progress bar output for atomic operation")
	}

	// Verify atomic completion message in stdout
	stdoutOutput := out.String()
	if !strings.Contains(stdoutOutput, "Atomically moved") {
		t.Error("Expected atomic completion message in stdout")
	}
}

func TestCopyCmd_DryRunWithProgress(t *testing.T) {
	tempDir := testutil.TempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test file
	testFile := filepath.Join(srcDir, "test.jpg")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	filesService := createTestFilesService(nil)
	dep := &deps.AppDeps{Files: filesService}
	cmd := createCopyCmd(dep)
	cmd.SetArgs([]string{"--dry-run", "--progress", "--overwrite", testFile, dstDir})

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("dry-run copy command with progress failed: %v", err)
	}

	// Verify dry-run output in stdout
	stdoutOutput := out.String()
	if !strings.Contains(stdoutOutput, "Would copy") {
		t.Error("Expected dry-run output in stdout")
	}

	// Progress should still be shown for dry-run
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "[") || !strings.Contains(stderrOutput, "]") {
		t.Error("Expected progress bar even during dry-run")
	}
	
	if !strings.Contains(stderrOutput, "✓") {
		t.Error("Expected completion checkmark even during dry-run")
	}
}