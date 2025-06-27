package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Tmunayyer/gocamelpack/files"
)

// collectSources expands a user-supplied path into absolute file paths.
//
// * file  → []{abs(file)}
// * dir   → []{abs(dir/entry1), abs(dir/entry2), …}
func collectSources(fs files.FilesService, userPath string) ([]string, error) {
	abs, err := filepath.Abs(userPath)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", userPath, err)
	}

	if fs.IsFile(abs) {
		return []string{abs}, nil
	}
	if fs.IsDirectory(abs) {
		entries, err := fs.ReadDirectory(abs)
		if err != nil {
			return nil, err
		}
		out := make([]string, len(entries))
		for i, e := range entries {
			out[i] = filepath.Join(abs, e)
		}
		return out, nil
	}
	return nil, fmt.Errorf("unknown src argument")
}

// destFromMetadata returns the destination path for a single source file.
func destFromMetadata(fs files.FilesService, src, dstRoot string) (string, error) {
	tags := fs.GetFileTags([]string{src})
	if len(tags) == 0 {
		return "", fmt.Errorf("no metadata for %s", src)
	}
	return fs.DestinationFromMetadata(tags[0], dstRoot)
}
