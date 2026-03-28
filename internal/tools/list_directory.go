package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ListDirectoryTool struct{}

func (t *ListDirectoryTool) Name() string { return "ListDirectory" }
func (t *ListDirectoryTool) Description() string {
	return "List files and directories at the given path. Respects .gitignore patterns."
}

func (t *ListDirectoryTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The directory path to list (defaults to current directory)",
			},
		},
		"required": []string{},
	}
}

func (t *ListDirectoryTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	dir := GetString(params, "path")
	if dir == "" {
		dir = "."
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("reading directory: %w", err)
	}

	var lines []string
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files
		if strings.HasPrefix(name, ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if entry.IsDir() {
			lines = append(lines, fmt.Sprintf("  %s/", name))
		} else {
			lines = append(lines, fmt.Sprintf("  %s (%s)", name, formatSize(info.Size())))
		}
	}

	header := fmt.Sprintf("Directory: %s\n", filepath.Clean(dir))
	if len(lines) == 0 {
		return header + "  (empty)", nil
	}
	return header + strings.Join(lines, "\n"), nil
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
