package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"amalgo/filter"
)

type Scanner struct {
	baseDir string
	filter  filter.Filter
}

func New(baseDir string, f filter.Filter) *Scanner {
	return &Scanner{
		baseDir: baseDir,
		filter:  f,
	}
}

func (s *Scanner) Scan() ([]string, error) {
	var files []string

	err := filepath.WalkDir(s.baseDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			fmt.Fprintf(os.Stderr, "warn: skipping %s: %v\n", path, walkErr)
			return nil
		}

		if !s.filter.ShouldInclude(path, d) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		ri := filter.RelPath(files[i], s.baseDir)
		rj := filter.RelPath(files[j], s.baseDir)
		return ri < rj
	})

	return files, nil
}
