package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type WriteFileTool struct{}

func (t *WriteFileTool) Name() string { return "WriteFile" }
func (t *WriteFileTool) Description() string {
	return "Write content to a file. Creates the file if it doesn't exist, or overwrites it if it does."
}

func (t *WriteFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path of the file to write",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path := GetString(params, "path")
	content := GetString(params, "content")
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	// Create parent directories if needed
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("creating directories: %w", err)
		}
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}
