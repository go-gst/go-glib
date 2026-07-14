package genmain

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// CleanGeneratedFiles removes all *.gen.go files from a given directory and then removes empty
// directories
func CleanGeneratedFiles(path string) error {
	abspath, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	dirfs := os.DirFS(abspath)

	var genfiles []string

	_ = fs.WalkDir(dirfs, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		filename := filepath.Base(name)

		if strings.HasSuffix(filename, ".gen.go") {
			genfiles = append(genfiles, name)
		}

		return nil
	})

	for _, f := range genfiles {
		abs := filepath.Join(abspath, f)
		err := os.Remove(abs)

		if err != nil {
			return err
		}

		slog.Info("removed file", "file", abs)
	}

	entries, err := os.ReadDir(abspath)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		subdirpath := filepath.Join(abspath, entry.Name())

		subdirEntries, err := os.ReadDir(subdirpath)

		if err != nil {
			return err
		}

		if len(subdirEntries) > 0 {
			continue
		}

		err = os.Remove(subdirpath)

		if err != nil {
			return err
		}
	}

	return nil
}
