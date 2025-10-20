package filter

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type HiddenFilter struct{}

func NewHiddenFilter() *HiddenFilter {
	return &HiddenFilter{}
}

func (h *HiddenFilter) ShouldInclude(path string, d fs.DirEntry) bool {
	name := filepath.Base(path)
	if strings.HasPrefix(name, ".") && name != "." && name != ".." {
		return false
	}
	return true
}
