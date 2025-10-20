package processor

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Path    string
	RelPath string
	Content []byte
	Ext     string
}

type Options struct {
	BaseDir       string
	HeadingLevel  int
	IncludeErrors bool
}

type Processor interface {
	Name() string

	FileExtension() string

	Process(files []FileInfo, opts Options) ([]byte, error)
}

type Registry struct {
	processors map[string]Processor
}

func NewRegistry() *Registry {
	return &Registry{
		processors: make(map[string]Processor),
	}
}

func (r *Registry) Register(p Processor) {
	r.processors[p.Name()] = p
}

func (r *Registry) Get(name string) (Processor, error) {
	p, ok := r.processors[name]
	if !ok {
		return nil, fmt.Errorf("unknown processor: %s", name)
	}
	return p, nil
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.processors))
	for name := range r.processors {
		names = append(names, name)
	}
	return names
}

func LoadFiles(paths []string, baseDir string) ([]FileInfo, error) {
	infos := make([]FileInfo, 0, len(paths))

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			infos = append(infos, FileInfo{
				Path:    path,
				RelPath: relPathOr(path, baseDir),
				Content: []byte(fmt.Sprintf("ERROR: could not read file: %v", err)),
				Ext:     filepath.Ext(path),
			})
			continue
		}

		infos = append(infos, FileInfo{
			Path:    path,
			RelPath: relPathOr(path, baseDir),
			Content: content,
			Ext:     filepath.Ext(path),
		})
	}

	return infos, nil
}

func relPathOr(path, base string) string {
	if rel, err := filepath.Rel(base, path); err == nil {
		return rel
	}
	return path
}
