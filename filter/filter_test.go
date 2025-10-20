package filter

import (
	"io/fs"
	"path/filepath"
	"testing"
)

func TestChain(t *testing.T) {
	t.Run("Empty chain allows all", func(t *testing.T) {
		chain := NewChain()
		mockEntry := &mockDirEntry{name: "test.go", isDir: false}

		if !chain.ShouldInclude("test.go", mockEntry) {
			t.Error("empty chain should allow all files")
		}
	})

	t.Run("Single filter", func(t *testing.T) {
		chain := NewChain(&alwaysFalseFilter{})
		mockEntry := &mockDirEntry{name: "test.go", isDir: false}

		if chain.ShouldInclude("test.go", mockEntry) {
			t.Error("chain should respect filter result")
		}
	})

	t.Run("Multiple filters AND logic", func(t *testing.T) {
		chain := NewChain(
			&alwaysTrueFilter{},
			&alwaysFalseFilter{},
		)
		mockEntry := &mockDirEntry{name: "test.go", isDir: false}

		// Should be false because one filter returns false
		if chain.ShouldInclude("test.go", mockEntry) {
			t.Error("chain should use AND logic")
		}
	})

	t.Run("Add filter", func(t *testing.T) {
		chain := NewChain()
		chain.Add(&alwaysFalseFilter{})

		mockEntry := &mockDirEntry{name: "test.go", isDir: false}
		if chain.ShouldInclude("test.go", mockEntry) {
			t.Error("added filter should be applied")
		}
	})
}

func TestBuildChain(t *testing.T) {
	t.Run("Build with extensions only", func(t *testing.T) {
		cfg := Config{
			Extensions: map[string]struct{}{".go": {}},
			BaseDir:    "/project",
		}

		chain, err := BuildChain(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if chain == nil {
			t.Fatal("expected chain to be created")
		}
	})

	t.Run("Build with hidden filter", func(t *testing.T) {
		cfg := Config{
			Extensions:    map[string]struct{}{".go": {}},
			IncludeHidden: false,
			BaseDir:       "/project",
		}

		chain, err := BuildChain(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Test that hidden files are filtered
		mockEntry := &mockDirEntry{name: ".hidden", isDir: false}
		if chain.ShouldInclude(".hidden", mockEntry) {
			t.Error("hidden filter should exclude hidden files")
		}
	})
	t.Run("Build with ignore dirs", func(t *testing.T) {
		cfg := Config{
			Extensions: map[string]struct{}{".go": {}, ".js": {}},
			IgnoreDirs: map[string]struct{}{"node_modules": {}, "vendor": {}},
			BaseDir:    "/project",
		}

		chain, err := BuildChain(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		t.Run("excludes the directory itself", func(t *testing.T) {
			path := "node_modules"
			mockEntry := &mockDirEntry{name: path, isDir: true}
			if chain.ShouldInclude(path, mockEntry) {
				t.Error("dir filter should exclude the ignored directory itself")
			}
		})

		t.Run("excludes files within an ignored directory", func(t *testing.T) {
			// filepath.Join will create "node_modules/some_package/index.js" on Linux/macOS
			// and "node_modules\some_package\index.js" on Windows.
			path := filepath.Join("node_modules", "some_package", "index.js")
			mockEntry := &mockDirEntry{name: "index.js", isDir: false}

			if chain.ShouldInclude(path, mockEntry) {
				t.Errorf("dir filter should exclude files within an ignored directory, path: %s", path)
			}
		})

		t.Run("includes files in other directories", func(t *testing.T) {
			path := filepath.Join("cmd", "app", "main.go")
			mockEntry := &mockDirEntry{name: "main.go", isDir: false}

			if !chain.ShouldInclude(path, mockEntry) {
				t.Errorf("chain should include files that are not in an ignored directory, path: %s", path)
			}
		})
	})
}

func TestRelPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		base     string
		expected string
	}{
		{
			name:     "simple relative",
			path:     filepath.Join("/home", "user", "project", "main.go"),
			base:     filepath.Join("/home", "user", "project"),
			expected: "main.go",
		},
		{
			name:     "nested relative",
			path:     filepath.Join("/home", "user", "project", "cmd", "root.go"),
			base:     filepath.Join("/home", "user", "project"),
			expected: filepath.Join("cmd", "root.go"),
		},
		{
			name:     "same path",
			path:     filepath.Join("/home", "user", "project"),
			base:     filepath.Join("/home", "user", "project"),
			expected: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RelPath(tt.path, tt.base)
			if filepath.ToSlash(result) != filepath.ToSlash(tt.expected) {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Mock filters for testing
type alwaysTrueFilter struct{}

func (f *alwaysTrueFilter) ShouldInclude(path string, d fs.DirEntry) bool {
	return true
}

type alwaysFalseFilter struct{}

func (f *alwaysFalseFilter) ShouldInclude(path string, d fs.DirEntry) bool {
	return false
}

// Mock DirEntry for testing
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() fs.FileMode          { return 0 }
func (m *mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }
