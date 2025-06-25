package main

import (
	"github.com/Tmunayyer/gocamelpack/cmd"
	"github.com/Tmunayyer/gocamelpack/deps"

	"github.com/Tmunayyer/gocamelpack/files"
)

func main() {
	files, _ := files.CreateFiles()
	defer files.Close()

	deps := &deps.AppDeps{Files: files}

	cmd.Execute(deps)
}
