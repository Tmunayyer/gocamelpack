package deps

import "github.com/Tmunayyer/gocamelpack/files"

type AppDeps struct {
	Files files.FilesService
	// Logger, Config, DB, etc.
}
