package girfiles_goglib

import (
	"embed"
)

//go:embed *.gir
var girFiles embed.FS

var GirFiles = ReadGirFiles(girFiles)

// ReadGirFiles reads all GIR files from the embedded filesystem and returns a
// map where the keys are the file names and the values are the file contents.
func ReadGirFiles(fs embed.FS) map[string][]byte {
	files := make(map[string][]byte)
	entries, err := fs.ReadDir(".")
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			data, err := fs.ReadFile(entry.Name())
			if err != nil {
				panic(err)
			}
			files[entry.Name()] = data
		}
	}

	return files
}
