package testutil

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var (
	once     sync.Once
	testRoot string
)

// TestRoot returns the shared temporary root directory.
// It is created exactly once for the entire 'go test' run.
func TestRoot(tb testing.TB) string {
	once.Do(func() {
		dir, err := os.MkdirTemp("", "gocamelpack_test_")
		if err != nil {
			tb.Fatalf("setup temp root: %v", err)
		}
		testRoot = dir

		// Clean up the whole tree at the end of the test run.
		// This fires only once because 'once' guarantees a single root.
		tb.Cleanup(func() { os.RemoveAll(testRoot) })
	})
	return testRoot
}

// TempDir returns a directory unique to the calling test,
// nested under the shared root, and guarantees it exists.
func TempDir(tb testing.TB) string {
	root := TestRoot(tb)
	dir := filepath.Join(root, tb.Name())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		tb.Fatalf("mkdir %s: %v", dir, err)
	}
	return dir
}
