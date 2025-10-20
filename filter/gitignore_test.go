package filter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitignoreFilter_CustomPatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		path     string
		isDir    bool
		expected bool
	}{
		{
			name:     "simple wildcard pattern",
			patterns: []string{"*.log"},
			path:     "/project/debug.log",
			isDir:    false,
			expected: false,
		},
		{
			name:     "directory pattern",
			patterns: []string{"build/"},
			path:     "/project/build",
			isDir:    true,
			expected: false,
		},
		{
			name:     "non-matching file",
			patterns: []string{"*.log"},
			path:     "/project/main.go",
			isDir:    false,
			expected: true,
		},
		{
			name:     "glob pattern",
			patterns: []string{"**/test/**"},
			path:     "/project/test/file.go",
			isDir:    false,
			expected: false,
		},
		{
			name:     "multiple patterns",
			patterns: []string{"*.log", "*.tmp", "build/"},
			path:     "/project/temp.tmp",
			isDir:    false,
			expected: false,
		},
		{
			name:     "empty pattern",
			patterns: []string{""},
			path:     "/project/file.go",
			isDir:    false,
			expected: true,
		},
		{
			name:     "comment pattern",
			patterns: []string{"# this is a comment", "*.log"},
			path:     "/project/debug.log",
			isDir:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewGitignoreFilter("/project", "", tt.patterns)
			if err != nil {
				t.Fatalf("failed to create filter: %v", err)
			}

			mockEntry := &mockDirEntry{name: filepath.Base(tt.path), isDir: tt.isDir}
			result := filter.ShouldInclude(tt.path, mockEntry)

			if result != tt.expected {
				t.Errorf("expected %v, got %v for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

func TestGitignoreFilter_NoPatterns(t *testing.T) {
	filter, err := NewGitignoreFilter("/project", "", []string{})
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	mockEntry := &mockDirEntry{name: "test.go", isDir: false}
	if !filter.ShouldInclude("/project/test.go", mockEntry) {
		t.Error("filter with no patterns should allow all files")
	}
}

func TestGitignoreFilter_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	content := `# Test gitignore
*.log
build/
temp/*
!important.log
`

	if err := os.WriteFile(gitignorePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .gitignore: %v", err)
	}

	filter, err := NewGitignoreFilter(tmpDir, gitignorePath, nil)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
	}{
		{
			name:     "ignored log file",
			path:     filepath.Join(tmpDir, "debug.log"),
			isDir:    false,
			expected: false,
		},
		{
			name:     "ignored build directory",
			path:     filepath.Join(tmpDir, "build"),
			isDir:    true,
			expected: false,
		},
		{
			name:     "allowed go file",
			path:     filepath.Join(tmpDir, "main.go"),
			isDir:    false,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEntry := &mockDirEntry{name: filepath.Base(tt.path), isDir: tt.isDir}
			result := filter.ShouldInclude(tt.path, mockEntry)

			if result != tt.expected {
				t.Errorf("expected %v, got %v for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

func TestGitignoreFilter_FileNotExists(t *testing.T) {
	filter, err := NewGitignoreFilter("/project", "/nonexistent/.gitignore", nil)
	if err != nil {
		t.Fatalf("should handle non-existent file gracefully: %v", err)
	}

	mockEntry := &mockDirEntry{name: "test.go", isDir: false}
	if !filter.ShouldInclude("/project/test.go", mockEntry) {
		t.Error("should allow files when gitignore doesn't exist")
	}
}

func TestAutoDetectGitignore(t *testing.T) {
	t.Run("gitignore exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")

		if err := os.WriteFile(gitignorePath, []byte("*.log"), 0644); err != nil {
			t.Fatalf("failed to create .gitignore: %v", err)
		}

		result := AutoDetectGitignore(tmpDir)
		if result != gitignorePath {
			t.Errorf("expected %s, got %s", gitignorePath, result)
		}
	})

	t.Run("gitignore does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		result := AutoDetectGitignore(tmpDir)
		if result != "" {
			t.Errorf("expected empty string, got %s", result)
		}
	})
}

func TestLoadGitignoreFile(t *testing.T) {
	t.Run("valid gitignore", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")

		content := `# Comment
*.log
build/

# Another comment
temp/*
`

		if err := os.WriteFile(gitignorePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write .gitignore: %v", err)
		}

		patterns, err := loadGitignoreFile(gitignorePath, tmpDir)
		if err != nil {
			t.Fatalf("failed to load .gitignore: %v", err)
		}

		if len(patterns) != 3 {
			t.Errorf("expected 3 patterns, got %d", len(patterns))
		}
	})

	t.Run("empty gitignore", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")

		if err := os.WriteFile(gitignorePath, []byte(""), 0644); err != nil {
			t.Fatalf("failed to write .gitignore: %v", err)
		}

		patterns, err := loadGitignoreFile(gitignorePath, tmpDir)
		if err != nil {
			t.Fatalf("failed to load .gitignore: %v", err)
		}

		if len(patterns) != 0 {
			t.Errorf("expected 0 patterns, got %d", len(patterns))
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		patterns, err := loadGitignoreFile("/nonexistent/.gitignore", "/base")
		if err != nil {
			t.Fatalf("should handle non-existent file: %v", err)
		}

		if patterns != nil {
			t.Error("expected nil patterns for non-existent file")
		}
	})
}
