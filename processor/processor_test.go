package processor

import (
	"path/filepath"
	"testing"
)

func TestRegistry(t *testing.T) {
	t.Run("Register and Get", func(t *testing.T) {
		reg := NewRegistry()
		mock := &mockProcessor{name: "test"}

		reg.Register(mock)

		proc, err := reg.Get("test")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if proc.Name() != "test" {
			t.Errorf("expected name 'test', got '%s'", proc.Name())
		}
	})

	t.Run("Get unknown processor", func(t *testing.T) {
		reg := NewRegistry()

		_, err := reg.Get("nonexistent")
		if err == nil {
			t.Fatal("expected error for unknown processor")
		}
	})

	t.Run("List processors", func(t *testing.T) {
		reg := NewRegistry()
		reg.Register(&mockProcessor{name: "proc1"})
		reg.Register(&mockProcessor{name: "proc2"})

		list := reg.List()
		if len(list) != 2 {
			t.Errorf("expected 2 processors, got %d", len(list))
		}
	})
}

func TestLoadFiles(t *testing.T) {
	t.Run("Empty file list", func(t *testing.T) {
		infos, err := LoadFiles([]string{}, "/base")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(infos) != 0 {
			t.Errorf("expected 0 files, got %d", len(infos))
		}
	})

	t.Run("Nonexistent file", func(t *testing.T) {
		infos, err := LoadFiles([]string{"/nonexistent/file.txt"}, "/base")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(infos) != 1 {
			t.Fatalf("expected 1 file info, got %d", len(infos))
		}

		content := string(infos[0].Content)
		if content == "" || len(content) < 5 {
			t.Error("expected error message in content")
		}
	})
}

func TestRelPathOr(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		base     string
		expected string
	}{
		{
			name:     "relative path success",
			path:     "/home/user/project/main.go",
			base:     "/home/user/project",
			expected: "main.go",
		},
		{
			name:     "nested relative path",
			path:     "/home/user/project/cmd/root.go",
			base:     "/home/user/project",
			expected: filepath.Join("cmd", "root.go"),
		},
		{
			name:     "same path",
			path:     "/home/user/project",
			base:     "/home/user/project",
			expected: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := relPathOr(tt.path, tt.base)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

type mockProcessor struct {
	name string
	ext  string
}

func (m *mockProcessor) Name() string {
	return m.name
}

func (m *mockProcessor) FileExtension() string {
	if m.ext == "" {
		return ".mock"
	}
	return m.ext
}

func (m *mockProcessor) Process(files []FileInfo, opts Options) ([]byte, error) {
	return []byte("mock output"), nil
}
