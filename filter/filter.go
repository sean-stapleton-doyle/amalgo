package filter

import (
	"io/fs"
	"path/filepath"
)

type Filter interface {
	ShouldInclude(path string, d fs.DirEntry) bool
}

type Chain struct {
	filters []Filter
}

func NewChain(filters ...Filter) *Chain {
	return &Chain{filters: filters}
}

func (c *Chain) Add(f Filter) {
	c.filters = append(c.filters, f)
}

func (c *Chain) ShouldInclude(path string, d fs.DirEntry) bool {
	for _, f := range c.filters {
		if !f.ShouldInclude(path, d) {
			return false
		}
	}
	return true
}

type Config struct {
	Extensions     map[string]struct{}
	IgnoreDirs     map[string]struct{}
	IncludeHidden  bool
	GitignorePath  string
	CustomPatterns []string
	BaseDir        string
}

func BuildChain(cfg Config) (*Chain, error) {
	chain := NewChain()

	if !cfg.IncludeHidden {
		chain.Add(NewHiddenFilter())
	}

	if len(cfg.IgnoreDirs) > 0 {
		chain.Add(NewDirFilter(cfg.IgnoreDirs))
	}

	if cfg.GitignorePath != "" || len(cfg.CustomPatterns) > 0 {
		gitFilter, err := NewGitignoreFilter(cfg.BaseDir, cfg.GitignorePath, cfg.CustomPatterns)
		if err != nil {
			return nil, err
		}
		chain.Add(gitFilter)
	}

	if len(cfg.Extensions) > 0 {
		chain.Add(NewExtensionFilter(cfg.Extensions))
	}

	return chain, nil
}

func RelPath(path, base string) string {
	if rel, err := filepath.Rel(base, path); err == nil {
		return rel
	}
	return path
}
