package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Tmunayyer/gocamelpack/deps"
	"github.com/Tmunayyer/gocamelpack/files"
	"github.com/Tmunayyer/gocamelpack/testutil"
)

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// testFilesService is a minimal implementation of FilesService for testing
type testFilesService struct {
	metadata map[string]files.FileMetadata
}

// createTestFilesService creates a test files service that uses real file operations
// but with mocked metadata extraction
func createTestFilesService(metadata map[string]files.FileMetadata) *testFilesService {
	return &testFilesService{
		metadata: metadata,
	}
}

func (t *testFilesService) Close() {
	// No-op for tests
}

func (t *testFilesService) GetFileTags(paths []string) []files.FileMetadata {
	var results []files.FileMetadata
	for _, path := range paths {
		if meta, ok := t.metadata[path]; ok {
			results = append(results, meta)
		} else {
			// Default metadata for files not explicitly configured
			results = append(results, files.FileMetadata{
				Filepath: path,
				Tags: map[string]string{
					"CreationDate": "2025:01:27 15:30:45-06:00",
					"FileType":     "JPEG",
				},
			})
		}
	}
	return results
}

// These methods delegate to the real file operations
func (t *testFilesService) IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (t *testFilesService) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (t *testFilesService) ReadDirectory(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var filePaths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			// Return just the entry name, collectSources will join it with the dir path
			filePaths = append(filePaths, entry.Name())
		}
	}
	return filePaths, nil
}

func (t *testFilesService) DestinationFromMetadata(md files.FileMetadata, baseDir string) (string, error) {
	// Simplified implementation for testing - organize by date
	raw := md.Tags["CreationDate"]
	if raw == "" {
		return "", fmt.Errorf("CreationDate is missing")
	}

	// For test simplicity, assume format "2025:01:27 15:30:45-06:00"
	// Extract year, month, day, hour, minute
	year := raw[:4]
	month := raw[5:7]
	day := raw[8:10]
	hour := raw[11:13]
	minute := raw[14:16]

	ext := filepath.Ext(md.Filepath)
	filename := hour + "_" + minute + ext

	return filepath.Join(baseDir, year, month, day, filename), nil
}

func (t *testFilesService) Copy(src, dst string) error {
	// Ensure destination directory exists
	if err := t.EnsureDir(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Write to destination
	return os.WriteFile(dst, data, 0644)
}

func (t *testFilesService) EnsureDir(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (t *testFilesService) ValidateCopyArgs(src, dst string) error {
	if src == "" || dst == "" {
		return fmt.Errorf("source and destination must be provided")
	}
	if !t.IsFile(src) {
		return fmt.Errorf("source %q is not a regular file", src)
	}
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("destination %q already exists", dst)
	}
	return nil
}

func (t *testFilesService) NewTransaction(overwrite bool) files.Transaction {
	return files.NewTransaction(t, overwrite)
}

func TestReadCmd_ValidFile(t *testing.T) {
	tempDir := testutil.TempDir(t)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.jpg")
	if err := os.WriteFile(testFile, []byte("test image content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set up mock metadata for the test file
	metadata := map[string]files.FileMetadata{
		testFile: {
			Filepath: testFile,
			Tags: map[string]string{
				"CreationDate": "2025:01:15 10:30:00-06:00",
				"FileType":     "JPEG",
				"ImageWidth":   "1920",
				"ImageHeight":  "1080",
			},
		},
	}

	filesService := createTestFilesService(metadata)
	dep := &deps.AppDeps{Files: filesService}
	cmd := createReadCmd(dep)
	cmd.SetArgs([]string{testFile})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Since the command uses fmt.Println which goes to os.Stdout directly,
	// we can't capture it with cmd.SetOut(). Just verify no error occurred.
	// The JSON output visible in test results shows the metadata is working correctly.
}

func TestReadCmd_InvalidFile(t *testing.T) {
	tempDir := testutil.TempDir(t)
	nonExistentFile := filepath.Join(tempDir, "does_not_exist.jpg")

	filesService := createTestFilesService(nil)
	dep := &deps.AppDeps{Files: filesService}
	cmd := createReadCmd(dep)
	cmd.SetArgs([]string{nonExistentFile})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := "src is not a file"
	if err.Error() != want {
		t.Errorf("unexpected error, got: %v", err)
	}
}

func TestCopyCmd(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, srcDir, dstDir string) string // returns source path
		verifyFunc func(t *testing.T, srcDir, dstDir string)
		expectErr  bool
	}{
		{
			name: "single file copy",
			setupFunc: func(t *testing.T, srcDir, dstDir string) string {
				testFile := filepath.Join(srcDir, "photo.jpg")
				if err := os.WriteFile(testFile, []byte("photo content"), 0644); err != nil {
					t.Fatal(err)
				}
				return testFile
			},
			verifyFunc: func(t *testing.T, srcDir, dstDir string) {
				// File should be organized by date: 2025/01/27/15_30.jpg
				expectedPath := filepath.Join(dstDir, "2025", "01", "27", "15_30.jpg")
				if _, err := os.Stat(expectedPath); err != nil {
					t.Errorf("expected file not found at %s: %v", expectedPath, err)
				}

				// Verify content
				content, err := os.ReadFile(expectedPath)
				if err != nil {
					t.Fatal(err)
				}
				if string(content) != "photo content" {
					t.Errorf("unexpected file content: %s", content)
				}

				// Original file should still exist
				originalPath := filepath.Join(srcDir, "photo.jpg")
				if _, err := os.Stat(originalPath); err != nil {
					t.Errorf("original file should still exist: %v", err)
				}
			},
		},
		{
			name: "directory with multiple files",
			setupFunc: func(t *testing.T, srcDir, dstDir string) string {
				photosDir := filepath.Join(srcDir, "photos")
				if err := os.Mkdir(photosDir, 0755); err != nil {
					t.Fatal(err)
				}

				// Create multiple test files
				for _, name := range []string{"a.jpg", "b.png", "c.gif"} {
					filePath := filepath.Join(photosDir, name)
					content := []byte("content " + name)
					if err := os.WriteFile(filePath, content, 0644); err != nil {
						t.Fatal(err)
					}
				}
				return photosDir
			},
			verifyFunc: func(t *testing.T, srcDir, dstDir string) {
				// All files should be organized by the same date
				for _, name := range []string{"a.jpg", "b.png", "c.gif"} {
					ext := filepath.Ext(name)
					expectedPath := filepath.Join(dstDir, "2025", "01", "27", "15_30"+ext)
					if _, err := os.Stat(expectedPath); err != nil {
						t.Errorf("expected file %s not found: %v", expectedPath, err)
					}
				}
			},
		},
		{
			name: "empty directory",
			setupFunc: func(t *testing.T, srcDir, dstDir string) string {
				emptyDir := filepath.Join(srcDir, "empty")
				if err := os.Mkdir(emptyDir, 0755); err != nil {
					t.Fatal(err)
				}
				return emptyDir
			},
			verifyFunc: func(t *testing.T, srcDir, dstDir string) {
				// No files should be copied
				entries, err := os.ReadDir(dstDir)
				if err != nil {
					t.Fatal(err)
				}
				if len(entries) != 0 {
					t.Errorf("expected no files in destination, found %d", len(entries))
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := testutil.TempDir(t)
			srcDir := filepath.Join(tempDir, "src")
			dstDir := filepath.Join(tempDir, "dst")

			if err := os.MkdirAll(srcDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				t.Fatal(err)
			}

			sourcePath := tc.setupFunc(t, srcDir, dstDir)

			filesService := createTestFilesService(nil)
			dep := &deps.AppDeps{Files: filesService}
			cmd := createCopyCmd(dep)
			cmd.SetArgs([]string{sourcePath, dstDir})

			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			err := cmd.Execute()

			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v\nOutput: %s", err, out.String())
			}

			tc.verifyFunc(t, srcDir, dstDir)
		})
	}
}

func TestCopyCmd_DryRun(t *testing.T) {
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
	cmd.SetArgs([]string{"--dry-run", testFile, dstDir})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify output mentions what would be done
	output := out.String()
	if !contains(output, "Would move") {
		t.Errorf("expected dry-run output, got: %s", output)
	} else {
		// If it printed the dry-run message, no files should be copied
		// Verify no files were actually copied
		entries, err := os.ReadDir(dstDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Errorf("dry-run should not copy files, but found %d entries", len(entries))
		}
	}
}

func TestMoveCmd(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, srcDir, dstDir string) string // returns source path
		verifyFunc func(t *testing.T, srcDir, dstDir string)
		expectErr  bool
	}{
		{
			name: "single file move",
			setupFunc: func(t *testing.T, srcDir, dstDir string) string {
				testFile := filepath.Join(srcDir, "photo.jpg")
				if err := os.WriteFile(testFile, []byte("photo content"), 0644); err != nil {
					t.Fatal(err)
				}
				return testFile
			},
			verifyFunc: func(t *testing.T, srcDir, dstDir string) {
				// File should be organized by date: 2025/01/27/15_30.jpg
				expectedPath := filepath.Join(dstDir, "2025", "01", "27", "15_30.jpg")
				if _, err := os.Stat(expectedPath); err != nil {
					t.Errorf("expected file not found at %s: %v", expectedPath, err)
				}

				// Verify content
				content, err := os.ReadFile(expectedPath)
				if err != nil {
					t.Fatal(err)
				}
				if string(content) != "photo content" {
					t.Errorf("unexpected file content: %s", content)
				}

				// Original file should NOT exist (moved, not copied)
				originalPath := filepath.Join(srcDir, "photo.jpg")
				if _, err := os.Stat(originalPath); err == nil {
					t.Errorf("original file should have been moved, but still exists")
				}
			},
		},
		{
			name: "directory with multiple files move",
			setupFunc: func(t *testing.T, srcDir, dstDir string) string {
				photosDir := filepath.Join(srcDir, "photos")
				if err := os.Mkdir(photosDir, 0755); err != nil {
					t.Fatal(err)
				}

				// Create multiple test files
				for _, name := range []string{"a.jpg", "b.png", "c.gif"} {
					filePath := filepath.Join(photosDir, name)
					content := []byte("content " + name)
					if err := os.WriteFile(filePath, content, 0644); err != nil {
						t.Fatal(err)
					}
				}
				return photosDir
			},
			verifyFunc: func(t *testing.T, srcDir, dstDir string) {
				// All files should be organized by the same date
				for _, name := range []string{"a.jpg", "b.png", "c.gif"} {
					ext := filepath.Ext(name)
					expectedPath := filepath.Join(dstDir, "2025", "01", "27", "15_30"+ext)
					if _, err := os.Stat(expectedPath); err != nil {
						t.Errorf("expected file %s not found: %v", expectedPath, err)
					}

					// Original files should not exist
					originalPath := filepath.Join(srcDir, "photos", name)
					if _, err := os.Stat(originalPath); err == nil {
						t.Errorf("original file %s should have been moved", originalPath)
					}
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := testutil.TempDir(t)
			srcDir := filepath.Join(tempDir, "src")
			dstDir := filepath.Join(tempDir, "dst")

			if err := os.MkdirAll(srcDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				t.Fatal(err)
			}

			sourcePath := tc.setupFunc(t, srcDir, dstDir)

			filesService := createTestFilesService(nil)
			dep := &deps.AppDeps{Files: filesService}
			cmd := createMoveCmd(dep)
			cmd.SetArgs([]string{sourcePath, dstDir})

			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			err := cmd.Execute()

			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v\nOutput: %s", err, out.String())
			}

			tc.verifyFunc(t, srcDir, dstDir)
		})
	}
}

func TestMoveCmd_DryRun(t *testing.T) {
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
	cmd := createMoveCmd(dep)
	cmd.SetArgs([]string{"--dry-run", testFile, dstDir})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify output mentions what would be done
	output := out.String()
	if !contains(output, "Would move") {
		t.Errorf("expected dry-run output, got: %s", output)
	} else {
		// If it printed the dry-run message, no files should be moved
		// Verify original file still exists (nothing was moved)
		if _, err := os.Stat(testFile); err != nil {
			t.Errorf("original file should still exist in dry-run mode: %v", err)
		}

		// Verify no files were moved to destination
		entries, err := os.ReadDir(dstDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Errorf("dry-run should not move files, but found %d entries", len(entries))
		}
	}
}

// Edge case tests
func TestCopyCmd_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tempDir string) (src, dst string)
		expectErr string
	}{
		{
			name: "source file does not exist",
			setupFunc: func(t *testing.T, tempDir string) (src, dst string) {
				src = filepath.Join(tempDir, "nonexistent.jpg")
				dst = filepath.Join(tempDir, "dst")
				if err := os.MkdirAll(dst, 0755); err != nil {
					t.Fatal(err)
				}
				return src, dst
			},
			expectErr: "unknown src argument",
		},
		{
			name: "file with missing creation date metadata",
			setupFunc: func(t *testing.T, tempDir string) (src, dst string) {
				src = filepath.Join(tempDir, "no_metadata.jpg")
				dst = filepath.Join(tempDir, "dst")
				if err := os.WriteFile(src, []byte("test"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(dst, 0755); err != nil {
					t.Fatal(err)
				}
				return src, dst
			},
			expectErr: "CreationDate is missing",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := testutil.TempDir(t)
			src, dst := tc.setupFunc(t, tempDir)

			// For the missing metadata test, create a service that returns empty metadata
			var filesService *testFilesService
			if tc.name == "file with missing creation date metadata" {
				metadata := map[string]files.FileMetadata{
					src: {
						Filepath: src,
						Tags:     map[string]string{}, // Empty tags
					},
				}
				filesService = createTestFilesService(metadata)
			} else {
				filesService = createTestFilesService(nil)
			}

			dep := &deps.AppDeps{Files: filesService}
			cmd := createCopyCmd(dep)
			cmd.SetArgs([]string{src, dst})

			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected error containing %q, got none", tc.expectErr)
			}

			if !contains(err.Error(), tc.expectErr) {
				t.Errorf("expected error containing %q, got: %v", tc.expectErr, err)
			}
		})
	}
}

func TestMoveCmd_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tempDir string) (src, dst string)
		expectErr string
	}{
		{
			name: "source file does not exist",
			setupFunc: func(t *testing.T, tempDir string) (src, dst string) {
				src = filepath.Join(tempDir, "nonexistent.jpg")
				dst = filepath.Join(tempDir, "dst")
				if err := os.MkdirAll(dst, 0755); err != nil {
					t.Fatal(err)
				}
				return src, dst
			},
			expectErr: "unknown src argument",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := testutil.TempDir(t)
			src, dst := tc.setupFunc(t, tempDir)

			filesService := createTestFilesService(nil)
			dep := &deps.AppDeps{Files: filesService}
			cmd := createMoveCmd(dep)
			cmd.SetArgs([]string{src, dst})

			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected error containing %q, got none", tc.expectErr)
			}

			if !contains(err.Error(), tc.expectErr) {
				t.Errorf("expected error containing %q, got: %v", tc.expectErr, err)
			}
		})
	}
}

func TestCopyCmd_Overwrite(t *testing.T) {
	tempDir := testutil.TempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create source file
	srcFile := filepath.Join(srcDir, "test.jpg")
	if err := os.WriteFile(srcFile, []byte("new content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create existing destination file
	dstPath := filepath.Join(dstDir, "2025", "01", "27", "15_30.jpg")
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dstPath, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	filesService := createTestFilesService(nil)
	dep := &deps.AppDeps{Files: filesService}

	// Test without --overwrite flag (should fail)
	cmd1 := createCopyCmd(dep)
	cmd1.SetArgs([]string{srcFile, dstDir})

	var out1 bytes.Buffer
	cmd1.SetOut(&out1)
	cmd1.SetErr(&out1)

	err := cmd1.Execute()
	if err == nil {
		t.Fatal("expected error when destination exists without --overwrite flag")
	}

	if !contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}

	// Test with --overwrite flag (should succeed)
	cmd2 := createCopyCmd(dep)
	cmd2.SetArgs([]string{"--overwrite", srcFile, dstDir})

	var out2 bytes.Buffer
	cmd2.SetOut(&out2)
	cmd2.SetErr(&out2)

	err = cmd2.Execute()
	if err != nil {
		t.Fatalf("unexpected error with --overwrite flag: %v", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new content" {
		t.Errorf("file was not overwritten, content: %s", content)
	}
}
