package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type EditFileTool struct{}

func (t *EditFileTool) Name() string { return "EditFile" }
func (t *EditFileTool) Description() string {
	return "Edit a file by replacing an exact string match with new content. The old_string must match exactly (including whitespace)."
}

func (t *EditFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path of the file to edit",
			},
			"old_string": map[string]any{
				"type":        "string",
				"description": "The exact string to find and replace",
			},
			"new_string": map[string]any{
				"type":        "string",
				"description": "The string to replace it with",
			},
		},
		"required": []string{"path", "old_string", "new_string"},
	}
}

func (t *EditFileTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path := GetString(params, "path")
	oldStr := GetString(params, "old_string")
	newStr := GetString(params, "new_string")

	if path == "" || oldStr == "" {
		return "", fmt.Errorf("path and old_string are required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	content := string(data)
	count := strings.Count(content, oldStr)
	if count == 0 {
		return "", fmt.Errorf("old_string not found in file")
	}
	if count > 1 {
		return "", fmt.Errorf("old_string found %d times, must be unique (found %d matches)", count, count)
	}

	newContent := strings.Replace(content, oldStr, newStr, 1)
	if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return fmt.Sprintf("Successfully replaced text in %s", path), nil
}
