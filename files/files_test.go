// files_test.go – unit tests for the Files helpers used by the copy command.
// Each test is hermetic: it spins up its own temporary directory via t.TempDir(),
// writes any scratch files it needs, and performs clean assertions without touching
// the user’s real filesystem.

package files

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Tmunayyer/gocamelpack/testutil"
)

// filePermRW represents rw-r--r-- (owner read/write; group and others read-only).
const filePermRW = 0o644

// filePermUserRW represents rw------- (owner read/write only).
const filePermUserRW = 0o600

// newFiles returns a minimal *Files instance that doesn’t invoke the real exiftool
// plumbing (pr: StdPath{} is a no‑op stub). This keeps tests fast and isolated.
func newFiles() *Files { return &Files{pr: StdPath{}} }

// TestEnsureDir verifies that EnsureDir creates a full nested directory path
// without error and that the resulting leaf is a directory on disk.
func TestEnsureDir(t *testing.T) {
	f := newFiles()
	tmp := testutil.TempDir(t)

	nested := filepath.Join(tmp, "a", "b", "c")
	if err := f.EnsureDir(nested, filePermRW); err != nil {
		t.Fatalf("EnsureDir returned error: %v", err)
	}
	if st, err := os.Stat(nested); err != nil || !st.IsDir() {
		t.Fatalf("directory %q was not created", nested)
	}
}

// TestValidateCopyArgs checks the basic argument validation logic:
//   - a valid source/destination passes
//   - an existing destination returns an error
func TestValidateCopyArgs(t *testing.T) {
	f := newFiles()
	tmp := testutil.TempDir(t)

	// --- happy path ---
	src := filepath.Join(tmp, "src.txt")
	if err := os.WriteFile(src, []byte("data"), filePermRW); err != nil {
		t.Fatalf("write src: %v", err)
	}

	dest := filepath.Join(tmp, "dest.txt")
	if err := f.ValidateCopyArgs(src, dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// --- error path: destination already exists ---
	if err := os.WriteFile(dest, nil, filePermRW); err != nil {
		t.Fatalf("prep dest: %v", err)
	}
	if err := f.ValidateCopyArgs(src, dest); err == nil {
		t.Fatalf("expected error for existing dest, got nil")
	}
}

// TestCopy performs an end‑to‑end single‑file copy and asserts:
//   - data integrity (byte‑perfect match)
//   - file mode bits are preserved
func TestCopy(t *testing.T) {
	f := newFiles()
	tmp := testutil.TempDir(t) // isolated playground

	// ----- arrange -----
	src := filepath.Join(tmp, "in.bin")
	want := []byte("shadowfax\n") // payload we expect after copy

	if err := os.WriteFile(src, want, filePermUserRW); err != nil {
		t.Fatalf("write src: %v", err)
	}

	dst := filepath.Join(tmp, "out.bin")
	if err := f.Copy(src, dst); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("content mismatch: got %q want %q", got, want)
	}

	// mode should match
	if srcInfo, _ := os.Stat(src); srcInfo != nil {
		if dstInfo, _ := os.Stat(dst); dstInfo.Mode() != srcInfo.Mode() {
			t.Fatalf("mode mismatch: src %v dst %v", srcInfo.Mode(), dstInfo.Mode())
		}
	}
}

// TestDestinationFromMetadata confirms that the helper constructs the expected
// YYYY/MM/DD/HH_mm path hierarchy from EXIF CreationDate metadata.
func TestDestinationFromMetadata(t *testing.T) {
	f := newFiles()
	md := FileMetadata{
		Tags: map[string]string{
			"CreationDate": "2025:01:27 07:31:15-06:00",
		},
	}
	base := "/media"
	got, err := f.DestinationFromMetadata(md, base)
	if err != nil {
		t.Fatalf("DestinationFromMetadata error: %v", err)
	}
	want := filepath.Join(base, "2025", "01", "27", "07_31") // matches helper’s format
	if got != want {
		t.Fatalf("path mismatch: got %q want %q", got, want)
	}
}

// TestDestinationFromMetadataExtension ensures that the destination path
// preserves the original file extension exactly (e.g., ".jpg" stays ".jpg").
func TestDestinationFromMetadataExtension(t *testing.T) {
	f := newFiles()

	md := FileMetadata{
		Filepath: "IMG_1234.jpg", // source file including extension
		Tags: map[string]string{
			"CreationDate": "2025:06:15 12:34:56-06:00",
		},
	}

	base := "/media"
	dst, err := f.DestinationFromMetadata(md, base)
	if err != nil {
		t.Fatalf("DestinationFromMetadata error: %v", err)
	}

	if !strings.HasSuffix(dst, ".jpg") {
		t.Fatalf("expected destination to keep .jpg extension, got %q", dst)
	}
}
