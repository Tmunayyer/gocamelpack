package files

import "path/filepath"

type PathResolver interface {
	Abs(path string) (string, error)
	Join(elem ...string) string
}

type StdPath struct{}

func (StdPath) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

func (StdPath) Join(elem ...string) string {
	return filepath.Join(elem...)
}
