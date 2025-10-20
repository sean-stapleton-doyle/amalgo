package cmd

import (
	"amalgo/processor"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessExtensions(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "single extension",
			input:     []string{".go"},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "multiple extensions",
			input:     []string{".go", ".rs", ".py"},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "comma separated",
			input:     []string{".go,.rs,.py"},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "mixed format",
			input:     []string{".go,.rs", ".py"},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "without leading dot",
			input:     []string{"go", "rs"},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "case insensitive",
			input:     []string{".GO", ".Rs", ".PY"},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "duplicate extensions",
			input:     []string{".go", ".go", ".go"},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "empty string",
			input:     []string{""},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "whitespace only",
			input:     []string{"  ", "\t"},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "with spaces",
			input:     []string{" .go ", " .rs "},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processExtensions(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != tt.wantCount {
				t.Errorf("expected %d extensions, got %d", tt.wantCount, len(result))
			}
		})
	}
}

func TestProcessIgnoreDirs(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		wantCount int
		contains  []string
	}{
		{
			name:      "single directory",
			input:     []string{"node_modules"},
			wantCount: 1,
			contains:  []string{"node_modules"},
		},
		{
			name:      "multiple directories",
			input:     []string{"node_modules", "vendor", ".git"},
			wantCount: 3,
			contains:  []string{"node_modules", "vendor", ".git"},
		},
		{
			name:      "with whitespace",
			input:     []string{" node_modules ", " vendor "},
			wantCount: 2,
			contains:  []string{"node_modules", "vendor"},
		},
		{
			name:      "empty strings filtered",
			input:     []string{"", "node_modules", "  ", "vendor"},
			wantCount: 2,
			contains:  []string{"node_modules", "vendor"},
		},
		{
			name:      "all empty",
			input:     []string{"", "  ", "\t"},
			wantCount: 0,
			contains:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processIgnoreDirs(tt.input)

			if len(result) != tt.wantCount {
				t.Errorf("expected %d directories, got %d", tt.wantCount, len(result))
			}

			for _, dir := range tt.contains {
				if _, ok := result[dir]; !ok {
					t.Errorf("expected result to contain '%s'", dir)
				}
			}
		})
	}
}

func TestHandleCommaSeparatedValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single value",
			input:    "value",
			expected: []string{"value"},
		},
		{
			name:     "multiple values",
			input:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with spaces",
			input:    " a , b , c ",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty parts",
			input:    "a,,b,,,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only commas",
			input:    ",,,",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleCommaSeparatedValues(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d values, got %d", len(tt.expected), len(result))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("at index %d: expected '%s', got '%s'", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestNormalizeExt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with leading dot",
			input:    ".go",
			expected: ".go",
		},
		{
			name:     "without leading dot",
			input:    "go",
			expected: ".go",
		},
		{
			name:     "uppercase",
			input:    ".GO",
			expected: ".go",
		},
		{
			name:     "mixed case",
			input:    ".Go",
			expected: ".go",
		},
		{
			name:     "with spaces",
			input:    "  .rs  ",
			expected: ".rs",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeExt(tt.input)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestWriteOutput(t *testing.T) {
	t.Run("write to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		outPath := filepath.Join(tmpDir, "output.md")

		content := []byte("# Test Content\n")
		err := writeOutput(content, outPath, 5)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed to read output file: %v", err)
		}

		if !bytes.Equal(data, content) {
			t.Errorf("file content mismatch")
		}
	})

	t.Run("write to stdout", func(t *testing.T) {

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		content := []byte("test content")
		err := writeOutput(content, "-", 1)

		os.Stdout = oldStdout

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = w.Close()

		if err != nil {
			t.Fatalf("Failed to close file: %v", err)
		}

		var buf bytes.Buffer
		_, err = buf.ReadFrom(r)

		if err != nil {
			t.Fatalf("Failed to read from buffer: %v", err)
		}

		if !bytes.Equal(buf.Bytes(), content) {
			t.Errorf("stdout content mismatch")
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		content := []byte("test")
		err := writeOutput(content, "/nonexistent/path/file.md", 1)

		if err == nil {
			t.Error("expected error for invalid path")
		}
	})
}

func TestRun_Integration(t *testing.T) {

	tmpDir := t.TempDir()

	files := map[string]string{
		"main.go":   "package main\n",
		"util.go":   "package util\n",
		"README.md": "# README\n",
		"test.txt":  "text file\n",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	t.Run("successful run", func(t *testing.T) {

		flagDir = tmpDir
		flagExts = []string{".go"}
		flagOut = filepath.Join(tmpDir, "output.md")
		flagIgnoreDirs = []string{}
		flagHeadingLevel = 1
		flagIncludeHidden = false
		flagFormat = "markdown"
		flagGitignore = ""
		flagUseGitignore = false
		flagIgnorePatterns = []string{}

		if registry == nil {
			registry = processor.NewRegistry()
			registry.Register(processor.NewMarkdownProcessor())
		}

		err := run(rootCmd, []string{})
		if err != nil {
			t.Fatalf("run failed: %v", err)
		}

		if _, err := os.Stat(flagOut); os.IsNotExist(err) {
			t.Error("output file was not created")
		}

		data, err := os.ReadFile(flagOut)
		if err != nil {
			t.Fatalf("failed to read output: %v", err)
		}

		output := string(data)
		if !strings.Contains(output, "main.go") {
			t.Error("output should contain main.go")
		}
		if !strings.Contains(output, "util.go") {
			t.Error("output should contain util.go")
		}
		if strings.Contains(output, "README.md") {
			t.Error("output should not contain README.md")
		}
	})

	t.Run("no matching files", func(t *testing.T) {
		flagDir = tmpDir
		flagExts = []string{".nonexistent"}
		flagOut = filepath.Join(tmpDir, "empty.md")
		flagIgnoreDirs = []string{}
		flagFormat = "markdown"
		flagGitignore = ""
		flagUseGitignore = false
		flagIgnorePatterns = []string{}

		err := run(rootCmd, []string{})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		flagDir = tmpDir
		flagExts = []string{".go"}
		flagFormat = "nonexistent"

		err := run(rootCmd, []string{})
		if err == nil {
			t.Error("expected error for invalid format")
		}
	})

	t.Run("invalid extension", func(t *testing.T) {
		flagDir = tmpDir
		flagExts = []string{"invalid"}
		flagFormat = "markdown"

		_, err := processExtensions(flagExts)
		if err == nil {
			t.Error("expected error for extension without dot")
		}
	})
}

func TestExecute(t *testing.T) {
	t.Run("command is initialized", func(t *testing.T) {
		if rootCmd == nil {
			t.Error("rootCmd should be initialized")
		}

		if rootCmd.Use != "amalgo" {
			t.Errorf("expected Use='amalgo', got '%s'", rootCmd.Use)
		}
	})
}
