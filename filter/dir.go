package filter

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type DirFilter struct {
	ignoreDirs map[string]struct{}
}

func NewDirFilter(ignoreDirs map[string]struct{}) *DirFilter {
	return &DirFilter{
		ignoreDirs: ignoreDirs,
	}
}

func (d *DirFilter) ShouldInclude(path string, entry fs.DirEntry) bool {
	normalizedPath := filepath.ToSlash(path)

	parts := strings.Split(normalizedPath, "/")

	if len(parts) > 0 {
		rootComponent := parts[0]
		if _, isIgnored := d.ignoreDirs[rootComponent]; isIgnored {
			return false
		}
	}

	return true
}
