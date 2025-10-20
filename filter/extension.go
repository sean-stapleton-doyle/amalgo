package filter

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type ExtensionFilter struct {
	extensions map[string]struct{}
}

func NewExtensionFilter(extensions map[string]struct{}) *ExtensionFilter {
	return &ExtensionFilter{
		extensions: extensions,
	}
}

func (e *ExtensionFilter) ShouldInclude(path string, d fs.DirEntry) bool {
	if d.IsDir() {
		return true
	}

	ext := strings.ToLower(filepath.Ext(path))
	_, ok := e.extensions[ext]
	return ok
}
