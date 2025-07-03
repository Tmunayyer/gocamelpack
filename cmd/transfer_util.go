package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Tmunayyer/gocamelpack/files"
	"github.com/Tmunayyer/gocamelpack/progress"
)

// collectSources expands a user-supplied path into absolute file paths.
//
// * file  → []{abs(file)}
// * dir   → []{abs(dir/entry1), abs(dir/entry2), …}
func collectSources(fs files.FilesService, userPath string) ([]string, error) {
	return collectSourcesWithProgress(fs, userPath, progress.NewNoOpReporter())
}

// collectSourcesWithProgress expands a user-supplied path into absolute file paths with progress reporting.
func collectSourcesWithProgress(fs files.FilesService, userPath string, reporter progress.ProgressReporter) ([]string, error) {
	abs, err := filepath.Abs(userPath)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", userPath, err)
	}

	if fs.IsFile(abs) {
		reporter.SetMessage("Collecting single file")
		reporter.SetTotal(1)
		reporter.SetCurrent(1)
		reporter.Finish()
		return []string{abs}, nil
	}
	if fs.IsDirectory(abs) {
		reporter.SetMessage("Reading directory")
		entries, err := fs.ReadDirectory(abs)
		if err != nil {
			return nil, err
		}
		
		reporter.SetMessage("Collecting files from directory")
		reporter.SetTotal(len(entries))
		
		out := make([]string, len(entries))
		for i, e := range entries {
			reporter.SetMessage(fmt.Sprintf("Collecting %s", e))
			out[i] = filepath.Join(abs, e)
			reporter.SetCurrent(i + 1)
		}
		reporter.Finish()
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
