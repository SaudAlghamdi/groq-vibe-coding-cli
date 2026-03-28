package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type SearchFilesTool struct{}

func (t *SearchFilesTool) Name() string { return "SearchFiles" }
func (t *SearchFilesTool) Description() string {
	return "Search for a pattern in files recursively. Returns matching lines with file paths and line numbers."
}

func (t *SearchFilesTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "The regex pattern to search for",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "The directory to search in (defaults to current directory)",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *SearchFilesTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	pattern := GetString(params, "pattern")
	searchPath := GetString(params, "path")
	if pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}
	if searchPath == "" {
		searchPath = "."
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid pattern: %w", err)
	}

	var results []string
	maxResults := 100

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if len(results) >= maxResults {
			return filepath.SkipAll
		}

		// Skip hidden dirs and common non-code dirs
		if info.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" || base == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip binary and large files
		if info.Size() > 1024*1024 {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if re.MatchString(line) {
				results = append(results, fmt.Sprintf("%s:%d: %s", path, lineNum, line))
				if len(results) >= maxResults {
					break
				}
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("search error: %w", err)
	}

	if len(results) == 0 {
		return "No matches found.", nil
	}

	result := strings.Join(results, "\n")
	if len(results) >= maxResults {
		result += fmt.Sprintf("\n... (truncated, showing first %d matches)", maxResults)
	}
	return result, nil
}
