package filter

import (
	"path/filepath"
	"testing"
)

func TestNewDirFilter(t *testing.T) {
	ignore := map[string]struct{}{"node_modules": {}}
	filter := NewDirFilter(ignore)

	if filter == nil {
		t.Fatal("NewDirFilter returned nil")
	}
	if _, ok := filter.ignoreDirs["node_modules"]; !ok {
		t.Error("ignoreDirs map was not initialized correctly")
	}
}

func TestDirFilter_ShouldInclude(t *testing.T) {

	ignoreDirs := map[string]struct{}{
		"node_modules": {},
		"vendor":       {},
		".git":         {},
	}
	filter := NewDirFilter(ignoreDirs)

	testCases := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
	}{

		{
			name:     "include file in root",
			path:     "main.go",
			isDir:    false,
			expected: true,
		},
		{
			name:     "include file in nested allowed directory",
			path:     filepath.Join("cmd", "app", "start.go"),
			isDir:    false,
			expected: true,
		},
		{
			name:     "include allowed directory",
			path:     "internal",
			isDir:    true,
			expected: true,
		},
		{
			name:     "include path that contains but does not start with ignored name",
			path:     filepath.Join("my_vendor_files", "main.go"),
			isDir:    false,
			expected: true,
		},

		{
			name:     "exclude ignored directory itself",
			path:     "node_modules",
			isDir:    true,
			expected: false,
		},
		{
			name:     "exclude file inside an ignored directory (Unix path)",
			path:     "node_modules/react/index.js",
			isDir:    false,
			expected: false,
		},
		{
			name:     "exclude nested directory inside an ignored directory",
			path:     "vendor/pkg/foo",
			isDir:    true,
			expected: false,
		},
		{
			name:     "exclude file in dotfile directory",
			path:     ".git/config",
			isDir:    false,
			expected: false,
		},

		{
			name:     "exclude file inside an ignored directory (Windows path)",
			path:     filepath.Join("node_modules", "react", "index.js"),
			isDir:    false,
			expected: false,
		},
		{
			name:     "exclude directory inside an ignored directory (Windows path)",
			path:     filepath.Join("vendor", "pkg", "foo"),
			isDir:    true,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockEntry := &mockDirEntry{
				name:  filepath.Base(tc.path),
				isDir: tc.isDir,
			}

			actual := filter.ShouldInclude(tc.path, mockEntry)

			if actual != tc.expected {
				t.Errorf("path: '%s', expected to be included: %v, but got: %v", tc.path, tc.expected, actual)
			}
		})
	}
}
