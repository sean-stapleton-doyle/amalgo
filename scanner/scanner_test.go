package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestScanner(t *testing.T) {
	t.Run("Scan directory with files", func(t *testing.T) {
		tmpDir := t.TempDir()

		files := []string{
			"main.go",
			"util.go",
			"cmd/root.go",
		}

		for _, f := range files {
			fullPath := filepath.Join(tmpDir, f)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Fatalf("failed to create directory: %v", err)
			}
			if err := os.WriteFile(fullPath, []byte("package main"), 0644); err != nil {
				t.Fatalf("failed to create file: %v", err)
			}
		}

		filter := &allowAllFilter{}
		scanner := New(tmpDir, filter)

		results, err := scanner.Scan()
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 files, got %d", len(results))
		}
	})

	t.Run("Scan with filtering", func(t *testing.T) {
		tmpDir := t.TempDir()

		files := map[string]string{
			"main.go":   "package main",
			"test.txt":  "text",
			"readme.md": "# README",
			"data.json": "{}",
		}

		for name, content := range files {
			if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
				t.Fatalf("failed to create file: %v", err)
			}
		}

		filter := &extensionOnlyFilter{ext: ".go"}
		scanner := New(tmpDir, filter)

		results, err := scanner.Scan()
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 file, got %d", len(results))
		}

		if len(results) > 0 && filepath.Base(results[0]) != "main.go" {
			t.Errorf("expected main.go, got %s", filepath.Base(results[0]))
		}
	})

	t.Run("Scan excludes directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		if err := os.MkdirAll(filepath.Join(tmpDir, "src"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		filter := &allowAllFilter{}
		scanner := New(tmpDir, filter)

		results, err := scanner.Scan()
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 file (no directories), got %d", len(results))
		}
	})

	t.Run("Scan empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		filter := &allowAllFilter{}
		scanner := New(tmpDir, filter)

		results, err := scanner.Scan()
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 files, got %d", len(results))
		}
	})

	t.Run("Scan with SkipDir", func(t *testing.T) {
		tmpDir := t.TempDir()

		if err := os.MkdirAll(filepath.Join(tmpDir, "ignore", "nested"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "ignore", "skip.go"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		filter := &skipDirFilter{skipDir: "ignore"}
		scanner := New(tmpDir, filter)

		results, err := scanner.Scan()
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 file, got %d", len(results))
		}
	})

	t.Run("Results are sorted", func(t *testing.T) {
		tmpDir := t.TempDir()

		files := []string{"z.go", "a.go", "m.go", "b.go"}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644); err != nil {
				t.Fatal(err)
			}
		}

		filter := &allowAllFilter{}
		scanner := New(tmpDir, filter)

		results, err := scanner.Scan()
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		expected := []string{"a.go", "b.go", "m.go", "z.go"}
		for i, f := range results {
			if filepath.Base(f) != expected[i] {
				t.Errorf("expected %s at position %d, got %s", expected[i], i, filepath.Base(f))
			}
		}
	})
}

type allowAllFilter struct{}

func (f *allowAllFilter) ShouldInclude(path string, d fs.DirEntry) bool {
	return true
}

type extensionOnlyFilter struct {
	ext string
}

func (f *extensionOnlyFilter) ShouldInclude(path string, d fs.DirEntry) bool {
	if d.IsDir() {
		return true
	}
	return filepath.Ext(path) == f.ext
}

type skipDirFilter struct {
	skipDir string
}

func (f *skipDirFilter) ShouldInclude(path string, d fs.DirEntry) bool {
	if d.IsDir() && filepath.Base(path) == f.skipDir {
		return false
	}
	return true
}
