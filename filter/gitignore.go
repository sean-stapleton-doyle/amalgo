package filter

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type GitignoreFilter struct {
	matcher gitignore.Matcher
	baseDir string
}

func NewGitignoreFilter(baseDir, gitignorePath string, customPatterns []string) (*GitignoreFilter, error) {
	var patterns []gitignore.Pattern

	if gitignorePath != "" {
		filePatterns, err := loadGitignoreFile(gitignorePath, baseDir)
		if err != nil {
			return nil, fmt.Errorf("loading gitignore file: %w", err)
		}
		patterns = append(patterns, filePatterns...)
	}

	for _, pattern := range customPatterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}
		patterns = append(patterns, gitignore.ParsePattern(pattern, nil))
	}

	if len(patterns) == 0 {
		return &GitignoreFilter{
			matcher: nil,
			baseDir: baseDir,
		}, nil
	}

	matcher := gitignore.NewMatcher(patterns)

	return &GitignoreFilter{
		matcher: matcher,
		baseDir: baseDir,
	}, nil
}

func (g *GitignoreFilter) ShouldInclude(path string, d fs.DirEntry) bool {
	if g.matcher == nil {
		return true
	}
	relPath := RelPath(path, g.baseDir)

	relPath = filepath.ToSlash(relPath)

	parts := strings.Split(relPath, "/")

	return !g.matcher.Match(parts, d.IsDir())
}

func loadGitignoreFile(path, baseDir string) ([]gitignore.Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var patterns []gitignore.Pattern
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pattern := gitignore.ParsePattern(line, nil)
		patterns = append(patterns, pattern)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

func AutoDetectGitignore(baseDir string) string {
	gitignorePath := filepath.Join(baseDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		return gitignorePath
	}
	return ""
}
