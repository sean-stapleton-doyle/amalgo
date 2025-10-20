package processor

import (
	"strings"
	"testing"
)

func TestMarkdownProcessor(t *testing.T) {
	proc := NewMarkdownProcessor()

	t.Run("Name and extension", func(t *testing.T) {
		if proc.Name() != "markdown" {
			t.Errorf("expected name 'markdown', got '%s'", proc.Name())
		}
		if proc.FileExtension() != ".md" {
			t.Errorf("expected extension '.md', got '%s'", proc.FileExtension())
		}
	})

	t.Run("Empty files", func(t *testing.T) {
		opts := Options{HeadingLevel: 1}
		result, err := proc.Process([]FileInfo{}, opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !strings.Contains(string(result), "No files found") {
			t.Error("expected 'No files found' message")
		}
	})

	t.Run("Single file", func(t *testing.T) {
		files := []FileInfo{
			{
				Path:    "/project/main.go",
				RelPath: "main.go",
				Content: []byte("package main\n"),
				Ext:     ".go",
			},
		}

		opts := Options{HeadingLevel: 1}
		result, err := proc.Process(files, opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		output := string(result)

		if !strings.Contains(output, "# main.go") {
			t.Error("expected heading with file path")
		}

		if !strings.Contains(output, "```go") {
			t.Error("expected go code block")
		}

		if !strings.Contains(output, "package main") {
			t.Error("expected file content")
		}
	})

	t.Run("Multiple files", func(t *testing.T) {
		files := []FileInfo{
			{
				RelPath: "main.go",
				Content: []byte("package main"),
				Ext:     ".go",
			},
			{
				RelPath: "util.go",
				Content: []byte("package util"),
				Ext:     ".go",
			},
		}

		opts := Options{HeadingLevel: 2}
		result, err := proc.Process(files, opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		output := string(result)

		if !strings.Contains(output, "## main.go") {
			t.Error("expected main.go heading")
		}
		if !strings.Contains(output, "## util.go") {
			t.Error("expected util.go heading")
		}
	})

	t.Run("Heading level clamping", func(t *testing.T) {
		files := []FileInfo{
			{RelPath: "test.go", Content: []byte("test"), Ext: ".go"},
		}

		tests := []struct {
			level    int
			expected string
		}{
			{0, "#"},
			{1, "#"},
			{3, "###"},
			{6, "######"},
			{10, "######"},
		}

		for _, tt := range tests {
			opts := Options{HeadingLevel: tt.level}
			result, err := proc.Process(files, opts)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !strings.Contains(string(result), tt.expected+" test.go") {
				t.Errorf("expected heading level '%s', got output: %s", tt.expected, string(result))
			}
		}
	})

	t.Run("Content without trailing newline", func(t *testing.T) {
		files := []FileInfo{
			{
				RelPath: "test.txt",
				Content: []byte("no newline"),
				Ext:     ".txt",
			},
		}

		opts := Options{HeadingLevel: 1}
		result, err := proc.Process(files, opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		output := string(result)

		if !strings.Contains(output, "no newline\n```") {
			t.Error("expected newline added before closing fence")
		}
	})
}

func TestInferLanguage(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".go", "go"},
		{".GO", "go"},
		{"go", "go"},
		{".rs", "rust"},
		{".py", "python"},
		{".js", "javascript"},
		{".ts", "typescript"},
		{".java", "java"},
		{".rb", "ruby"},
		{".cpp", "cpp"},
		{".html", "html"},
		{".json", "json"},
		{".yaml", "yaml"},
		{".yml", "yaml"},
		{".md", "markdown"},
		{".unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := inferLanguage(tt.ext)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		lo       int
		hi       int
		expected int
	}{
		{"below range", 0, 1, 6, 1},
		{"in range", 3, 1, 6, 3},
		{"above range", 10, 1, 6, 6},
		{"at lower bound", 1, 1, 6, 1},
		{"at upper bound", 6, 1, 6, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clamp(tt.value, tt.lo, tt.hi)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}
