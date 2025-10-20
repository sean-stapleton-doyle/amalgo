package filter

import (
	"testing"
)

func TestHiddenFilter(t *testing.T) {
	filter := NewHiddenFilter()

	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
	}{
		{
			name:     "regular file",
			path:     "/project/main.go",
			isDir:    false,
			expected: true,
		},
		{
			name:     "hidden file",
			path:     "/project/.gitignore",
			isDir:    false,
			expected: false,
		},
		{
			name:     "hidden directory",
			path:     "/project/.git",
			isDir:    true,
			expected: false,
		},
		{
			name:     "regular directory",
			path:     "/project/src",
			isDir:    true,
			expected: true,
		},
		{
			name:     "file starting with dot in name",
			path:     ".hidden.txt",
			isDir:    false,
			expected: false,
		},
		{
			name:     "current directory",
			path:     ".",
			isDir:    true,
			expected: true,
		},
		{
			name:     "parent directory",
			path:     "..",
			isDir:    true,
			expected: true,
		},
		{
			name:     "nested hidden file",
			path:     "/project/src/.config",
			isDir:    false,
			expected: false,
		},
		{
			name:     "file with dot in middle",
			path:     "/project/main.test.go",
			isDir:    false,
			expected: true,
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
