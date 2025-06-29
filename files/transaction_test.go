package files

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTransaction_CopyOperations tests basic transaction functionality with copy operations.
func TestTransaction_CopyOperations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")
	
	// Create source directory and files
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	
	// Create test files
	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "file2.txt")
	
	if err := os.WriteFile(file1, []byte("content1"), 0o644); err != nil {
		t.Fatalf("Failed to create test file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0o644); err != nil {
		t.Fatalf("Failed to create test file2: %v", err)
	}
	
	// Create destination paths
	dst1 := filepath.Join(dstDir, "file1.txt")
	dst2 := filepath.Join(dstDir, "file2.txt")
	
	// Create a Files instance for testing
	files := newFiles()
	
	// Create a transaction
	tx := NewTransaction(files, false)
	
	// Add operations
	if err := tx.AddCopy(file1, dst1); err != nil {
		t.Fatalf("Failed to add copy operation: %v", err)
	}
	if err := tx.AddCopy(file2, dst2); err != nil {
		t.Fatalf("Failed to add copy operation: %v", err)
	}
	
	// Validate operations
	if err := tx.Validate(); err != nil {
		t.Fatalf("Transaction validation failed: %v", err)
	}
	
	// Execute transaction
	if err := tx.Execute(); err != nil {
		t.Fatalf("Transaction execution failed: %v", err)
	}
	
	// Verify files were copied
	if _, err := os.Stat(dst1); os.IsNotExist(err) {
		t.Errorf("File1 was not copied to destination")
	}
	if _, err := os.Stat(dst2); os.IsNotExist(err) {
		t.Errorf("File2 was not copied to destination")
	}
	
	// Verify content
	content1, err := os.ReadFile(dst1)
	if err != nil {
		t.Fatalf("Failed to read copied file1: %v", err)
	}
	if string(content1) != "content1" {
		t.Errorf("File1 content mismatch: got %q, want %q", string(content1), "content1")
	}
	
	content2, err := os.ReadFile(dst2)
	if err != nil {
		t.Fatalf("Failed to read copied file2: %v", err)
	}
	if string(content2) != "content2" {
		t.Errorf("File2 content mismatch: got %q, want %q", string(content2), "content2")
	}
}

// TestTransaction_MoveOperations tests transaction functionality with move operations.
func TestTransaction_MoveOperations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")
	
	// Create source directory and files
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	
	// Create test files
	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "file2.txt")
	
	if err := os.WriteFile(file1, []byte("content1"), 0o644); err != nil {
		t.Fatalf("Failed to create test file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0o644); err != nil {
		t.Fatalf("Failed to create test file2: %v", err)
	}
	
	// Create destination paths
	dst1 := filepath.Join(dstDir, "file1.txt")
	dst2 := filepath.Join(dstDir, "file2.txt")
	
	// Create a Files instance for testing
	files := newFiles()
	
	// Create a transaction
	tx := NewTransaction(files, false)
	
	// Add operations
	if err := tx.AddMove(file1, dst1); err != nil {
		t.Fatalf("Failed to add move operation: %v", err)
	}
	if err := tx.AddMove(file2, dst2); err != nil {
		t.Fatalf("Failed to add move operation: %v", err)
	}
	
	// Validate operations
	if err := tx.Validate(); err != nil {
		t.Fatalf("Transaction validation failed: %v", err)
	}
	
	// Execute transaction
	if err := tx.Execute(); err != nil {
		t.Fatalf("Transaction execution failed: %v", err)
	}
	
	// Verify files were moved (exist at destination, not at source)
	if _, err := os.Stat(dst1); os.IsNotExist(err) {
		t.Errorf("File1 was not moved to destination")
	}
	if _, err := os.Stat(dst2); os.IsNotExist(err) {
		t.Errorf("File2 was not moved to destination")
	}
	
	// Verify source files no longer exist
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Errorf("File1 still exists at source after move")
	}
	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Errorf("File2 still exists at source after move")
	}
}

// TestTransaction_Rollback tests that failed transactions roll back correctly.
func TestTransaction_Rollback(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")
	
	// Create source directory and files
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	
	// Create test files
	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "file2.txt")
	
	if err := os.WriteFile(file1, []byte("content1"), 0o644); err != nil {
		t.Fatalf("Failed to create test file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0o644); err != nil {
		t.Fatalf("Failed to create test file2: %v", err)
	}
	
	// Create destination paths
	dst1 := filepath.Join(dstDir, "file1.txt")
	dst2 := filepath.Join(dstDir, "file2.txt")
	
	// Create a conflicting file at dst2 to cause the second copy to fail
	if err := os.WriteFile(dst2, []byte("existing"), 0o644); err != nil {
		t.Fatalf("Failed to create conflicting file: %v", err)
	}
	
	// Create a Files instance for testing
	files := newFiles()
	
	// Create a transaction
	tx := NewTransaction(files, false) // overwrite=false, so dst2 conflict will fail
	
	// Add operations
	if err := tx.AddCopy(file1, dst1); err != nil {
		t.Fatalf("Failed to add copy operation: %v", err)
	}
	if err := tx.AddCopy(file2, dst2); err != nil {
		t.Fatalf("Failed to add copy operation: %v", err)
	}
	
	// Execute transaction - this should fail and rollback
	err := tx.Execute()
	if err == nil {
		t.Fatalf("Expected transaction to fail due to conflicting destination")
	}
	
	// Verify rollback occurred - dst1 should not exist
	if _, err := os.Stat(dst1); !os.IsNotExist(err) {
		t.Errorf("File1 should have been rolled back but still exists at destination")
	}
	
	// Verify dst2 still has original content
	content, err := os.ReadFile(dst2)
	if err != nil {
		t.Fatalf("Failed to read dst2: %v", err)
	}
	if string(content) != "existing" {
		t.Errorf("dst2 content changed: got %q, want %q", string(content), "existing")
	}
}

// TestTransaction_Operations tests that the Operations() method returns correct operations.
func TestTransaction_Operations(t *testing.T) {
	files := newFiles()
	tx := NewTransaction(files, false)
	
	// Add some operations
	err := tx.AddCopy("/src/file1.txt", "/dst/file1.txt")
	if err != nil {
		t.Fatalf("Failed to add copy operation: %v", err)
	}
	
	err = tx.AddMove("/src/file2.txt", "/dst/file2.txt")
	if err != nil {
		t.Fatalf("Failed to add move operation: %v", err)
	}
	
	// Get operations
	ops := tx.Operations()
	
	if len(ops) != 2 {
		t.Errorf("Expected 2 operations, got %d", len(ops))
	}
	
	// Check first operation
	if ops[0].Type() != OperationCopy {
		t.Errorf("Expected first operation to be copy, got %v", ops[0].Type())
	}
	if ops[0].Source() != "/src/file1.txt" {
		t.Errorf("Expected source /src/file1.txt, got %s", ops[0].Source())
	}
	if ops[0].Destination() != "/dst/file1.txt" {
		t.Errorf("Expected destination /dst/file1.txt, got %s", ops[0].Destination())
	}
	
	// Check second operation
	if ops[1].Type() != OperationMove {
		t.Errorf("Expected second operation to be move, got %v", ops[1].Type())
	}
	if ops[1].Source() != "/src/file2.txt" {
		t.Errorf("Expected source /src/file2.txt, got %s", ops[1].Source())
	}
	if ops[1].Destination() != "/dst/file2.txt" {
		t.Errorf("Expected destination /dst/file2.txt, got %s", ops[1].Destination())
	}
}

// TestTransaction_Validation tests transaction validation.
func TestTransaction_Validation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")
	
	// Create source and destination directories
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	
	// Create test file
	file1 := filepath.Join(srcDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create Files instance
	files := newFiles()
	
	// Test validation with non-existent source
	tx := NewTransaction(files, false)
	err := tx.AddCopy("/nonexistent/file.txt", filepath.Join(dstDir, "file.txt"))
	if err != nil {
		t.Fatalf("Failed to add copy operation: %v", err)
	}
	
	err = tx.Validate()
	if err == nil {
		t.Errorf("Expected validation to fail for non-existent source")
	}
	
	// Test validation with existing destination
	dst1 := filepath.Join(dstDir, "existing.txt")
	if err := os.WriteFile(dst1, []byte("existing"), 0o644); err != nil {
		t.Fatalf("Failed to create existing destination file: %v", err)
	}
	
	tx2 := NewTransaction(files, false)
	err = tx2.AddCopy(file1, dst1)
	if err != nil {
		t.Fatalf("Failed to add copy operation: %v", err)
	}
	
	err = tx2.Validate()
	if err == nil {
		t.Errorf("Expected validation to fail for existing destination")
	}
	
	// Test validation with overwrite enabled
	tx3 := NewTransaction(files, true)
	err = tx3.AddCopy(file1, dst1)
	if err != nil {
		t.Fatalf("Failed to add copy operation: %v", err)
	}
	
	err = tx3.Validate()
	if err != nil {
		t.Errorf("Expected validation to succeed with overwrite enabled: %v", err)
	}
}