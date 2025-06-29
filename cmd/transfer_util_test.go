package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Tmunayyer/gocamelpack/files"
)

// ----- minimal mock that satisfies files.FilesService -----
type utilMock struct {
	isFile       func(string) bool
	isDir        func(string) bool
	readDir      func(string) ([]string, error)
	getTags      func([]string) []files.FileMetadata
	destFromMeta func(files.FileMetadata, string) (string, error)
}

func (m utilMock) Close()                                       {}
func (m utilMock) Copy(_, _ string) error                       { return nil }
func (m utilMock) EnsureDir(_ string, _ os.FileMode) error      { return nil }
func (m utilMock) ValidateCopyArgs(_, _ string) error           { return nil }
func (m utilMock) IsFile(p string) bool                         { return m.isFile(p) }
func (m utilMock) IsDirectory(p string) bool                    { return m.isDir(p) }
func (m utilMock) ReadDirectory(p string) ([]string, error)     { return m.readDir(p) }
func (m utilMock) GetFileTags(ps []string) []files.FileMetadata { return m.getTags(ps) }
func (m utilMock) DestinationFromMetadata(md files.FileMetadata, base string) (string, error) {
	return m.destFromMeta(md, base)
}
func (m utilMock) NewTransaction(overwrite bool) files.Transaction {
	return files.NewTransaction(m, overwrite)
}

// -----------------------------------------------------------

func TestCollectSources_File(t *testing.T) {
	path := "example.jpg"
	abs, _ := filepath.Abs(path)

	mock := utilMock{
		isFile: func(p string) bool { return strings.HasSuffix(p, path) },
		isDir:  func(string) bool { return false },
	}

	got, err := collectSources(mock, path)
	if err != nil {
		t.Fatalf("collectSources: %v", err)
	}
	if !reflect.DeepEqual(got, []string{abs}) {
		t.Fatalf("want %v, got %v", []string{abs}, got)
	}
}

func TestCollectSources_Directory(t *testing.T) {
	dir := "photos"
	entries := []string{"a.png", "b.jpg"}
	mock := utilMock{
		isFile: func(string) bool { return false },
		isDir:  func(p string) bool { return strings.HasSuffix(p, dir) },
		readDir: func(p string) ([]string, error) {
			return entries, nil
		},
	}

	got, err := collectSources(mock, dir)
	if err != nil {
		t.Fatalf("collectSources dir: %v", err)
	}

	wantAbs, _ := filepath.Abs(dir)
	want := []string{filepath.Join(wantAbs, "a.png"), filepath.Join(wantAbs, "b.jpg")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestDestFromMetadata(t *testing.T) {
	src := "IMG_0001.jpg"
	mock := utilMock{
		getTags: func(ps []string) []files.FileMetadata {
			return []files.FileMetadata{{Filepath: src}}
		},
		destFromMeta: func(md files.FileMetadata, base string) (string, error) {
			return filepath.Join(base, "2025", "06", "27", "00_00.jpg"), nil
		},
	}

	dst, err := destFromMetadata(mock, src, "/media")
	if err != nil {
		t.Fatalf("destFromMetadata: %v", err)
	}
	if !strings.HasSuffix(dst, "00_00.jpg") {
		t.Fatalf("unexpected dst: %s", dst)
	}
}
