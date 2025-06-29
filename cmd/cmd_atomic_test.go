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

func TestCopyCmd_Atomic(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, srcDir, dstDir string) (string, map[string]files.FileMetadata)
		verifyFunc func(t *testing.T, srcDir, dstDir string, failed bool)
		overwrite  bool
	}{
		{
			name: "successful atomic copy of directory with multiple files",
			setupFunc: func(t *testing.T, srcDir, dstDir string) (string, map[string]files.FileMetadata) {
				photosDir := filepath.Join(srcDir, "photos")
				if err := os.Mkdir(photosDir, 0755); err != nil {
					t.Fatal(err)
				}

				metadata := make(map[string]files.FileMetadata)

				// Create multiple test files with different dates
				testFiles := []struct {
					name string
					date string
				}{
					{"photo1.jpg", "2025:01:01 10:00:00-06:00"},
					{"photo2.jpg", "2025:01:02 11:00:00-06:00"},
					{"photo3.jpg", "2025:01:03 12:00:00-06:00"},
				}

				for _, f := range testFiles {
					path := filepath.Join(photosDir, f.name)
					if err := os.WriteFile(path, []byte("content "+f.name), 0644); err != nil {
						t.Fatal(err)
					}
					metadata[path] = files.FileMetadata{
						Filepath: path,
						Tags: map[string]string{
							"CreationDate": f.date,
							"FileType":     "JPEG",
						},
					}
				}

				return photosDir, metadata
			},
			verifyFunc: func(t *testing.T, srcDir, dstDir string, failed bool) {
				if failed {
					t.Error("transaction should have succeeded")
				}

				// Verify all files were copied to their respective destinations
				expectedFiles := []string{
					filepath.Join(dstDir, "2025", "01", "01", "10_00.jpg"),
					filepath.Join(dstDir, "2025", "01", "02", "11_00.jpg"),
					filepath.Join(dstDir, "2025", "01", "03", "12_00.jpg"),
				}

				for _, path := range expectedFiles {
					if _, err := os.Stat(path); err != nil {
						t.Errorf("expected file not found: %s", path)
					}
				}
			},
		},
		{
			name: "atomic copy with rollback on conflict",
			setupFunc: func(t *testing.T, srcDir, dstDir string) (string, map[string]files.FileMetadata) {
				photosDir := filepath.Join(srcDir, "photos")
				if err := os.Mkdir(photosDir, 0755); err != nil {
					t.Fatal(err)
				}

				metadata := make(map[string]files.FileMetadata)

				// Create test files
				for i := 1; i <= 3; i++ {
					name := fmt.Sprintf("photo%d.jpg", i)
					path := filepath.Join(photosDir, name)
					if err := os.WriteFile(path, []byte("content "+name), 0644); err != nil {
						t.Fatal(err)
					}
					metadata[path] = files.FileMetadata{
						Filepath: path,
						Tags: map[string]string{
							"CreationDate": "2025:01:01 10:00:00-06:00", // Same date for all
							"FileType":     "JPEG",
						},
					}
				}

				// Create a conflicting file at the destination
				conflictPath := filepath.Join(dstDir, "2025", "01", "01", "10_00.jpg")
				if err := os.MkdirAll(filepath.Dir(conflictPath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(conflictPath, []byte("existing content"), 0644); err != nil {
					t.Fatal(err)
				}

				return photosDir, metadata
			},
			verifyFunc: func(t *testing.T, srcDir, dstDir string, failed bool) {
				if !failed {
					t.Error("transaction should have failed due to conflict")
				}

				// Verify NO new files were created (transaction rolled back)
				// The only file should be the pre-existing conflict file
				walkCount := 0
				filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
					if err == nil && !info.IsDir() {
						walkCount++
						// Verify it's the original conflict file
						content, _ := os.ReadFile(path)
						if string(content) != "existing content" {
							t.Errorf("unexpected file content: %s", content)
						}
					}
					return nil
				})

				if walkCount != 1 {
					t.Errorf("expected only 1 file (the conflict), found %d", walkCount)
				}
			},
		},
		{
			name:      "atomic copy with --overwrite succeeds despite conflicts",
			overwrite: true,
			setupFunc: func(t *testing.T, srcDir, dstDir string) (string, map[string]files.FileMetadata) {
				photosDir := filepath.Join(srcDir, "photos")
				if err := os.Mkdir(photosDir, 0755); err != nil {
					t.Fatal(err)
				}

				metadata := make(map[string]files.FileMetadata)

				// Create test file
				name := "photo1.jpg"
				path := filepath.Join(photosDir, name)
				if err := os.WriteFile(path, []byte("new content "+name), 0644); err != nil {
					t.Fatal(err)
				}
				metadata[path] = files.FileMetadata{
					Filepath: path,
					Tags: map[string]string{
						"CreationDate": "2025:01:01 10:00:00-06:00",
						"FileType":     "JPEG",
					},
				}

				// Create a conflicting file
				conflictPath := filepath.Join(dstDir, "2025", "01", "01", "10_00.jpg")
				if err := os.MkdirAll(filepath.Dir(conflictPath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(conflictPath, []byte("old content"), 0644); err != nil {
					t.Fatal(err)
				}

				return photosDir, metadata
			},
			verifyFunc: func(t *testing.T, srcDir, dstDir string, failed bool) {
				if failed {
					t.Error("transaction should have succeeded with --overwrite")
				}

				// Verify the file was overwritten
				path := filepath.Join(dstDir, "2025", "01", "01", "10_00.jpg")
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if string(content) != "new content photo1.jpg" {
					t.Errorf("file was not overwritten correctly: %s", content)
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

			sourcePath, metadata := tc.setupFunc(t, srcDir, dstDir)

			filesService := createTestFilesService(metadata)
			dep := &deps.AppDeps{Files: filesService}
			cmd := createCopyCmd(dep)

			args := []string{"--atomic"}
			if tc.overwrite {
				args = append(args, "--overwrite")
			}
			args = append(args, sourcePath, dstDir)
			cmd.SetArgs(args)

			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			err := cmd.Execute()
			failed := err != nil

			tc.verifyFunc(t, srcDir, dstDir, failed)
		})
	}
}

func TestMoveCmd_Atomic(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, srcDir, dstDir string) (string, map[string]files.FileMetadata)
		verifyFunc func(t *testing.T, srcDir, dstDir string, failed bool)
	}{
		{
			name: "successful atomic move of directory with multiple files",
			setupFunc: func(t *testing.T, srcDir, dstDir string) (string, map[string]files.FileMetadata) {
				photosDir := filepath.Join(srcDir, "photos")
				if err := os.Mkdir(photosDir, 0755); err != nil {
					t.Fatal(err)
				}

				metadata := make(map[string]files.FileMetadata)

				// Create multiple test files
				testFiles := []struct {
					name string
					date string
				}{
					{"photo1.jpg", "2025:02:01 10:00:00-06:00"},
					{"photo2.jpg", "2025:02:02 11:00:00-06:00"},
					{"photo3.jpg", "2025:02:03 12:00:00-06:00"},
				}

				for _, f := range testFiles {
					path := filepath.Join(photosDir, f.name)
					if err := os.WriteFile(path, []byte("content "+f.name), 0644); err != nil {
						t.Fatal(err)
					}
					metadata[path] = files.FileMetadata{
						Filepath: path,
						Tags: map[string]string{
							"CreationDate": f.date,
							"FileType":     "JPEG",
						},
					}
				}

				return photosDir, metadata
			},
			verifyFunc: func(t *testing.T, srcDir, dstDir string, failed bool) {
				if failed {
					t.Error("transaction should have succeeded")
				}

				// Verify all files were moved to their destinations
				expectedFiles := []string{
					filepath.Join(dstDir, "2025", "02", "01", "10_00.jpg"),
					filepath.Join(dstDir, "2025", "02", "02", "11_00.jpg"),
					filepath.Join(dstDir, "2025", "02", "03", "12_00.jpg"),
				}

				for _, path := range expectedFiles {
					if _, err := os.Stat(path); err != nil {
						t.Errorf("expected file not found: %s", path)
					}
				}

				// Verify source files were removed
				photosDir := filepath.Join(srcDir, "photos")
				for i := 1; i <= 3; i++ {
					srcPath := filepath.Join(photosDir, fmt.Sprintf("photo%d.jpg", i))
					if _, err := os.Stat(srcPath); err == nil {
						t.Errorf("source file should have been moved: %s", srcPath)
					}
				}
			},
		},
		{
			name: "atomic move with rollback on conflict",
			setupFunc: func(t *testing.T, srcDir, dstDir string) (string, map[string]files.FileMetadata) {
				photosDir := filepath.Join(srcDir, "photos")
				if err := os.Mkdir(photosDir, 0755); err != nil {
					t.Fatal(err)
				}

				metadata := make(map[string]files.FileMetadata)

				// Create test files
				for i := 1; i <= 3; i++ {
					name := fmt.Sprintf("photo%d.jpg", i)
					path := filepath.Join(photosDir, name)
					if err := os.WriteFile(path, []byte("content "+name), 0644); err != nil {
						t.Fatal(err)
					}
					metadata[path] = files.FileMetadata{
						Filepath: path,
						Tags: map[string]string{
							"CreationDate": "2025:02:01 10:00:00-06:00", // Same date
							"FileType":     "JPEG",
						},
					}
				}

				// Create a conflicting file
				conflictPath := filepath.Join(dstDir, "2025", "02", "01", "10_00.jpg")
				if err := os.MkdirAll(filepath.Dir(conflictPath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(conflictPath, []byte("existing content"), 0644); err != nil {
					t.Fatal(err)
				}

				return photosDir, metadata
			},
			verifyFunc: func(t *testing.T, srcDir, dstDir string, failed bool) {
				if !failed {
					t.Error("transaction should have failed due to conflict")
				}

				// Verify all source files still exist (rolled back)
				photosDir := filepath.Join(srcDir, "photos")
				for i := 1; i <= 3; i++ {
					srcPath := filepath.Join(photosDir, fmt.Sprintf("photo%d.jpg", i))
					if _, err := os.Stat(srcPath); err != nil {
						t.Errorf("source file should still exist after rollback: %s", srcPath)
					}
				}

				// Verify only the conflict file exists in destination
				walkCount := 0
				filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
					if err == nil && !info.IsDir() {
						walkCount++
					}
					return nil
				})

				if walkCount != 1 {
					t.Errorf("expected only 1 file in destination, found %d", walkCount)
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

			sourcePath, metadata := tc.setupFunc(t, srcDir, dstDir)

			filesService := createTestFilesService(metadata)
			dep := &deps.AppDeps{Files: filesService}
			cmd := createMoveCmd(dep)

			args := []string{"--atomic", sourcePath, dstDir}
			cmd.SetArgs(args)

			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			err := cmd.Execute()
			failed := err != nil

			tc.verifyFunc(t, srcDir, dstDir, failed)
		})
	}
}

func TestAtomicCmd_DryRun(t *testing.T) {
	tempDir := testutil.TempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a directory with test files
	photosDir := filepath.Join(srcDir, "photos")
	if err := os.Mkdir(photosDir, 0755); err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 3; i++ {
		path := filepath.Join(photosDir, fmt.Sprintf("photo%d.jpg", i))
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	filesService := createTestFilesService(nil)
	dep := &deps.AppDeps{Files: filesService}
	cmd := createCopyCmd(dep)

	cmd.SetArgs([]string{"--atomic", "--dry-run", photosDir, dstDir})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify output shows what would be done
	output := out.String()
	if !contains(output, "Would") {
		t.Errorf("expected dry-run output, got: %s", output)
	}

	// Verify no files were actually copied
	entries, err := os.ReadDir(dstDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dry-run should not copy files, but found %d entries", len(entries))
	}
}