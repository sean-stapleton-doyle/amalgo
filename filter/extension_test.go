package filter

import (
	"testing"
)

func TestExtensionFilter(t *testing.T) {
	extensions := map[string]struct{}{
		".go": {},
		".rs": {},
		".py": {},
	}

	filter := NewExtensionFilter(extensions)

	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
	}{
		{
			name:     "matching extension .go",
			path:     "/project/main.go",
			isDir:    false,
			expected: true,
		},
		{
			name:     "matching extension .rs",
			path:     "/project/lib.rs",
			isDir:    false,
			expected: true,
		},
		{
			name:     "matching extension .py",
			path:     "/project/script.py",
			isDir:    false,
			expected: true,
		},
		{
			name:     "non-matching extension",
			path:     "/project/readme.md",
			isDir:    false,
			expected: false,
		},
		{
			name:     "directory always included",
			path:     "/project/src",
			isDir:    true,
			expected: true,
		},
		{
			name:     "file with no extension",
			path:     "/project/Makefile",
			isDir:    false,
			expected: false,
		},
		{
			name:     "uppercase extension",
			path:     "/project/MAIN.GO",
			isDir:    false,
			expected: true,
		},
		{
			name:     "mixed case extension",
			path:     "/project/test.Go",
			isDir:    false,
			expected: true,
		},
		{
			name:     "multiple dots",
			path:     "/project/test.min.js",
			isDir:    false,
			expected: false,
		},
		{
			name:     "file starting with dot",
			path:     "/project/.gitignore",
			isDir:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEntry := &mockDirEntry{name: tt.path, isDir: tt.isDir}
			result := filter.ShouldInclude(tt.path, mockEntry)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtensionFilter_EmptyExtensions(t *testing.T) {
	filter := NewExtensionFilter(map[string]struct{}{})

	mockDir := &mockDirEntry{name: "src", isDir: true}
	if !filter.ShouldInclude("src", mockDir) {
		t.Error("empty extensions should still allow directories")
	}

	mockFile := &mockDirEntry{name: "main.go", isDir: false}
	if filter.ShouldInclude("main.go", mockFile) {
		t.Error("empty extensions should exclude all files")
	}
}
