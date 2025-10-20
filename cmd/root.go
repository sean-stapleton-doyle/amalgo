package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"amalgo/filter"
	"amalgo/processor"
	"amalgo/scanner"

	"github.com/spf13/cobra"
)

var (
	flagDir            string
	flagExts           []string
	flagOut            string
	flagIgnoreDirs     []string
	flagHeadingLevel   int
	flagIncludeHidden  bool
	flagFormat         string
	flagGitignore      string
	flagUseGitignore   bool
	flagIgnorePatterns []string
)

var (
	registry *processor.Registry
)

var rootCmd = &cobra.Command{
	Use:   "amalgo",
	Short: "Concatenate files by extension into a single output file.",
	Long: `Amalgo recursively scans a directory for files with given extension(s)
and produces an amalgamated markdown file.
Supports .gitignore patterns for flexible file filtering.
Handy for passing a small project as context to LLMs or for documentation.`,
	RunE: run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	registry = processor.NewRegistry()
	registry.Register(processor.NewMarkdownProcessor())

	formats := strings.Join(registry.List(), ", ")

	rootCmd.Flags().StringVarP(&flagDir, "dir", "d", ".", "Root directory to scan")
	rootCmd.Flags().StringSliceVarP(&flagExts, "ext", "e", nil, "File extension(s) to include (e.g. .rs,.py or repeat -e) [required]")
	rootCmd.Flags().StringVarP(&flagOut, "out", "o", "", "Output file path (use '-' for stdout, default: concat.<format>)")
	rootCmd.Flags().StringSliceVarP(&flagIgnoreDirs, "ignore-dirs", "i", []string{".git", "node_modules", "vendor"}, "Directory names to ignore")
	rootCmd.Flags().IntVarP(&flagHeadingLevel, "heading-level", "l", 1, "Markdown heading level (1-6)")
	rootCmd.Flags().BoolVar(&flagIncludeHidden, "include-hidden", false, "Include hidden files and directories")
	rootCmd.Flags().StringVarP(&flagFormat, "format", "f", "markdown", fmt.Sprintf("Output format: %s", formats))
	rootCmd.Flags().StringVarP(&flagGitignore, "gitignore", "g", "", "Path to .gitignore file (default: auto-detect in base dir)")
	rootCmd.Flags().BoolVar(&flagUseGitignore, "use-gitignore", true, "Automatically use .gitignore in base directory if present")
	rootCmd.Flags().StringSliceVarP(&flagIgnorePatterns, "ignore-pattern", "p", nil, "Custom gitignore-style patterns to exclude (can be repeated)")

	_ = rootCmd.MarkFlagRequired("ext")
}

func run(cmd *cobra.Command, args []string) error {
	proc, err := registry.Get(flagFormat)
	if err != nil {
		return fmt.Errorf("%w\nAvailable formats: %s", err, strings.Join(registry.List(), ", "))
	}

	extSet, err := processExtensions(flagExts)
	if err != nil {
		return err
	}

	ignoreSet := processIgnoreDirs(flagIgnoreDirs)
	baseDir := filepath.Clean(flagDir)

	gitignorePath := flagGitignore
	if gitignorePath == "" && flagUseGitignore {
		gitignorePath = filter.AutoDetectGitignore(baseDir)
		if gitignorePath != "" {
			fmt.Fprintf(os.Stderr, "Using .gitignore: %s\n", gitignorePath)
		}
	}

	filterCfg := filter.Config{
		Extensions:     extSet,
		IgnoreDirs:     ignoreSet,
		IncludeHidden:  flagIncludeHidden,
		GitignorePath:  gitignorePath,
		CustomPatterns: flagIgnorePatterns,
		BaseDir:        baseDir,
	}

	filterChain, err := filter.BuildChain(filterCfg)
	if err != nil {
		return fmt.Errorf("building filter chain: %w", err)
	}

	s := scanner.New(baseDir, filterChain)
	files, err := s.Scan()
	if err != nil {
		return fmt.Errorf("scanning files: %w", err)
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No files found matching criteria")
		return nil
	}

	fileInfos, err := processor.LoadFiles(files, baseDir)
	if err != nil {
		return fmt.Errorf("loading files: %w", err)
	}

	opts := processor.Options{
		BaseDir:      baseDir,
		HeadingLevel: flagHeadingLevel,
	}
	content, err := proc.Process(fileInfos, opts)
	if err != nil {
		return fmt.Errorf("processing files: %w", err)
	}

	outPath := flagOut
	if outPath == "" {
		outPath = "concat" + proc.FileExtension()
	}

	if err := writeOutput(content, outPath, len(files)); err != nil {
		return err
	}

	return nil
}

func processExtensions(rawExts []string) (map[string]struct{}, error) {
	extSet := make(map[string]struct{})
	for _, raw := range rawExts {
		for _, tok := range handleCommaSeparatedValues(raw) {
			ext := normalizeExt(tok)
			if ext == "" || !strings.HasPrefix(ext, ".") {
				return nil, errors.New("each --ext must start with a dot, e.g. .rs or .py")
			}
			extSet[strings.ToLower(ext)] = struct{}{}
		}
	}
	if len(extSet) == 0 {
		return nil, errors.New("no valid extensions provided")
	}
	return extSet, nil
}

func processIgnoreDirs(rawIgnore []string) map[string]struct{} {
	ignoreSet := make(map[string]struct{}, len(rawIgnore))
	for _, d := range rawIgnore {
		d = strings.TrimSpace(d)
		if d != "" {
			ignoreSet[d] = struct{}{}
		}
	}
	return ignoreSet
}

func writeOutput(content []byte, outPath string, fileCount int) error {
	if outPath == "-" {
		_, err := os.Stdout.Write(content)
		return err
	}

	if err := os.WriteFile(outPath, content, 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Wrote %d file(s) to %s\n", fileCount, outPath)
	return nil
}

func handleCommaSeparatedValues(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func normalizeExt(e string) string {
	e = strings.TrimSpace(e)
	if e == "" {
		return ""
	}
	if !strings.HasPrefix(e, ".") {
		e = "." + e
	}
	return strings.ToLower(e)
}
