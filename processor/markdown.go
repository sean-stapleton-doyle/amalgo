package processor

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

type MarkdownProcessor struct{}

func NewMarkdownProcessor() *MarkdownProcessor {
	return &MarkdownProcessor{}
}

func (m *MarkdownProcessor) Name() string {
	return "markdown"
}

func (m *MarkdownProcessor) FileExtension() string {
	return ".md"
}

func (m *MarkdownProcessor) Process(files []FileInfo, opts Options) ([]byte, error) {
	var out bytes.Buffer

	if len(files) == 0 {
		fmt.Fprintln(&out, "_No files found._")
		return out.Bytes(), nil
	}

	headingLevel := clamp(opts.HeadingLevel, 1, 6)
	heading := strings.Repeat("#", headingLevel)

	for _, file := range files {
		relPath := filepath.ToSlash(file.RelPath)

		fmt.Fprintf(&out, "%s %s\n", heading, relPath)

		lang := inferLanguage(file.Ext)
		fmt.Fprintf(&out, "```%s\n", lang)

		out.Write(file.Content)

		if len(file.Content) == 0 || file.Content[len(file.Content)-1] != '\n' {
			out.WriteByte('\n')
		}

		out.WriteString("```\n\n")
	}

	return out.Bytes(), nil
}

func inferLanguage(ext string) string {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	langMap := map[string]string{
		".go":   "go",
		".rs":   "rust",
		".py":   "python",
		".js":   "javascript",
		".ts":   "typescript",
		".java": "java",
		".c":    "c",
		".cpp":  "cpp",
		".cs":   "csharp",
		".rb":   "ruby",
		".php":  "php",
		".sh":   "bash",
		".bash": "bash",
		".zsh":  "bash",
		".html": "html",
		".css":  "css",
		".scss": "scss",
		".json": "json",
		".yaml": "yaml",
		".yml":  "yaml",
		".xml":  "xml",
		".toml": "toml",
		".sql":  "sql",
		".md":   "markdown",
		".txt":  "text",
	}

	if lang, ok := langMap[ext]; ok {
		return lang
	}

	return ""
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
