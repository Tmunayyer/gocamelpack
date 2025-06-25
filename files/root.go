package files

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
)

type FileMetadata struct {
	Filepath string
	Tags     map[string]string
}

type FilesService interface {
	Close()
	IsFile(path string) bool
	IsDirectory(path string) bool
	GetFileTags(paths []string) []FileMetadata
	ReadDirectory(dirPath string) ([]string, error)
	DestinationFromMetadata(tags FileMetadata, baseDir string) (string, error)
	Copy(src, dst string) error
	EnsureDir(path string, perm os.FileMode) error
	ValidateCopyArgs(src, dst string) error
}

type Files struct {
	et *exiftool.Exiftool
	pr PathResolver
}

func CreateFiles() (*Files, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return nil, fmt.Errorf("error intializing exiftool: %v", err)
	}

	f := Files{
		et: et,
		pr: StdPath{},
	}

	return &f, nil
}

func (f *Files) GetFileTags(files []string) []FileMetadata {
	raw := f.et.ExtractMetadata(files...)
	var result []FileMetadata
	for _, r := range raw {
		tags := make(map[string]string)
		for k, v := range r.Fields {
			tags[k] = fmt.Sprintf("%v", v)
		}
		result = append(result, FileMetadata{
			Filepath: r.File,
			Tags:     tags,
		})
	}
	return result
}

func (f *Files) Close() {
	f.et.Close()
}

func (f *Files) IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false // File does not exist or other error
	}
	return !info.IsDir()
}

func (f *Files) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (f *Files) ReadDirectory(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var filePaths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			absPath, err := f.pr.Abs(f.pr.Join(dirPath, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to get absolute path: %w", err)
			}
			filePaths = append(filePaths, absPath)
		}
	}

	return filePaths, nil
}

func (f *Files) DestinationFromMetadata(md FileMetadata, baseDir string) (string, error) {
	raw := md.Tags["CreationDate"]
	if raw == "" {
		return "", fmt.Errorf("CreationDate is missing")
	}

	// Normalize to RFC3339-like format
	// From: 2025:01:27 07:31:15-06:00
	// To:   2025-01-27T07:31:15-06:00
	rfcish := strings.Replace(raw, ":", "-", 2)
	rfcish = strings.Replace(rfcish, " ", "T", 1)

	t, err := time.Parse(time.RFC3339, rfcish)
	if err != nil {
		return "", fmt.Errorf("failed to parse CreationDate %q: %w", raw, err)
	}

	year, month, day := t.Date()
	hour := fmt.Sprintf("%02d", t.Hour())
	minute := fmt.Sprintf("%02d", t.Minute())

	dir := filepath.Join(
		baseDir,
		fmt.Sprintf("%04d", year),
		fmt.Sprintf("%02d", int(month)),
		fmt.Sprintf("%02d", day),
	)

	filename := fmt.Sprintf("%s_%s", hour, minute)

	return filepath.Join(dir, filename), nil
}

// EnsureDir creates the directory path (and parents) with the provided permissions.
func (f *Files) EnsureDir(path string, perm os.FileMode) error {
	if path == "" {
		return fmt.Errorf("directory path is empty")
	}
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("creating directory %q: %w", path, err)
	}
	return nil
}

// ValidateCopyArgs performs basic sanity checks before copy begins.
func (f *Files) ValidateCopyArgs(src, dst string) error {
	if src == "" || dst == "" {
		return fmt.Errorf("source and destination must be provided")
	}
	if !f.IsFile(src) {
		return fmt.Errorf("source %q is not a regular file", src)
	}
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("destination %q already exists", dst)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking destination: %w", err)
	}
	return nil
}

// Copy performs a single‑threaded, safe file copy preserving permissions.
func (f *Files) Copy(src, dst string) error {
	// Basic validations
	if err := f.ValidateCopyArgs(src, dst); err != nil {
		return err
	}

	// Open source
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %q: %w", src, err)
	}
	defer in.Close()

	srcInfo, err := in.Stat()
	if err != nil {
		return fmt.Errorf("stat %q: %w", src, err)
	}

	// Ensure destination directory exists
	if err := f.EnsureDir(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	// Create destination exclusively so we never clobber existing files
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_EXCL, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("create %q: %w", dst, err)
	}

	var copyErr error
	defer func() {
		if copyErr != nil {
			out.Close()
			os.Remove(dst)
		}
	}()

	// Transfer data
	if _, copyErr = io.Copy(out, in); copyErr != nil {
		return fmt.Errorf("copy data: %w", copyErr)
	}

	// Flush to disk
	if err = out.Sync(); err != nil {
		out.Close()
		os.Remove(dst)
		return fmt.Errorf("sync %q: %w", dst, err)
	}

	return out.Close()
}
