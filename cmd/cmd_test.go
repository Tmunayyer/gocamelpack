package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Tmunayyer/gocamelpack/deps"
	"github.com/Tmunayyer/gocamelpack/files"
)

type mockFiles struct {
	close                     func()
	isFileFn                  func(string) bool
	isDirectoryFn             func(string) bool
	getFileTagsFn             func([]string) []files.FileMetadata
	calledWith                []string
	readDirectoryFn           func(string) ([]string, error)
	destinationFromMetadataFn func(files.FileMetadata, string) (string, error)
}

func (m *mockFiles) Close() {
	m.close()
}

func (m *mockFiles) IsFile(path string) bool {
	return m.isFileFn(path)
}

func (m *mockFiles) IsDirectory(path string) bool {
	return m.isDirectoryFn(path)
}

func (m *mockFiles) GetFileTags(paths []string) []files.FileMetadata {
	m.calledWith = append(m.calledWith, paths...)
	if m.getFileTagsFn != nil {
		return m.getFileTagsFn(paths)
	}

	mockTags := make(map[string]string)

	return []files.FileMetadata{
		{
			Filepath: "default/test.file",
			Tags:     mockTags,
		},
	}
}

func (m *mockFiles) ReadDirectory(dirPath string) ([]string, error) {
	if m.readDirectoryFn != nil {
		return m.readDirectoryFn(dirPath)
	}

	return []string{
		"default/test/file/one.txt",
		"default/test/file/two.txt",
	}, nil
}

func (m *mockFiles) DestinationFromMetadata(tags files.FileMetadata, base string) (string, error) {
	if m.destinationFromMetadataFn != nil {
		return m.destinationFromMetadataFn(tags, base)
	}
	return "default/path", nil
}

func TestReadCmd_ValidFile(t *testing.T) {
	mock := &mockFiles{
		isFileFn: func(path string) bool {
			return true
		},
	}
	dep := &deps.AppDeps{Files: mock}
	cmd := createReadCmd(dep)
	cmd.SetArgs([]string{"image.jpg"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mock.calledWith) != 1 || mock.calledWith[0] != "image.jpg" {
		t.Errorf("GetFileTags not called correctly: %v", mock.calledWith)
	}
}

func TestReadCmd_InvalidFile(t *testing.T) {
	mock := &mockFiles{
		isFileFn: func(path string) bool {
			return false
		},
	}
	dep := &deps.AppDeps{Files: mock}
	cmd := createReadCmd(dep)
	cmd.SetArgs([]string{"notafile.jpg"})

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
		name                      string
		path                      string
		isDir                     bool
		readDirRes                []string
		readDirErr                error
		expectErr                 bool
		expectCalls               []string
		destinationFromMetadataFn func(files.FileMetadata, string) (string, error)
	}{
		{
			name:        "single file input",
			path:        "singular.txt",
			isDir:       false,
			expectErr:   false,
			expectCalls: []string{"abs/photos/singular.txt"},
			destinationFromMetadataFn: func(tags files.FileMetadata, base string) (string, error) {
				return fmt.Sprintf("mocked/%s", files.FileMetadata{Filepath: "something/singular.txt"}), nil
			},
		},
		{
			name:        "empty directory",
			path:        "mydir",
			isDir:       true,
			readDirRes:  []string{},
			expectErr:   false,
			expectCalls: []string{},
			destinationFromMetadataFn: func(tags files.FileMetadata, base string) (string, error) {
				return fmt.Sprintf("mocked/%s", files.FileMetadata{Filepath: "something/singular.txt"}), nil
			},
		},
		{
			name:       "directory with files",
			path:       "photos",
			isDir:      true,
			readDirRes: []string{"a.png", "b.jpg"},
			expectErr:  false,
			// should forward absolute paths of "a.png", "b.jpg"
			expectCalls: []string{"abs/photos/a.png", "abs/photos/b.jpg"},
			destinationFromMetadataFn: func(tags files.FileMetadata, base string) (string, error) {
				return fmt.Sprintf("mocked/%s", files.FileMetadata{Filepath: "something/singular.txt"}), nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockFiles{
				isFileFn: func(p string) bool {
					return !tc.isDir
				},
				isDirectoryFn: func(p string) bool {
					return p == tc.path && tc.isDir
				},
				readDirectoryFn: func(dir string) ([]string, error) {
					if !tc.isDir || dir != tc.path {
						t.Errorf("unexpected ReadDirectory call: %s", dir)
					}
					return tc.readDirRes, tc.readDirErr
				},
				getFileTagsFn:             nil,
				destinationFromMetadataFn: tc.destinationFromMetadataFn,
			}

			dep := &deps.AppDeps{Files: mock}
			cmd := createCopyCmd(dep)
			cmd.SetArgs([]string{tc.path, "outout"})

			buf := bytes.NewBuffer(nil)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// TODO
			// 1. reenable the below code, then make cmd work
			// 2. test for actually copying files
			// 		2a. make copying files work

			// Check that GetFileTags was called for each absolute file
			if len(mock.calledWith) != len(tc.expectCalls) {
				t.Fatalf("expected %d calls to GetFileTags, got %d", len(tc.expectCalls), len(mock.calledWith))
			}
			// for i, cw := range mock.calledWith {
			// 	if cw != tc.expectCalls[i] {
			// 		t.Errorf("call %d: expected %q, got %q", i, tc.expectCalls[i], cw)
			// 	}
			// }
		})
	}
}
