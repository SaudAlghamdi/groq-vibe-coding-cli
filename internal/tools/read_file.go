package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type ReadFileTool struct{}

func (t *ReadFileTool) Name() string { return "ReadFile" }
func (t *ReadFileTool) Description() string {
	return "Read the contents of a file. Returns the file content with line numbers."
}

func (t *ReadFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path of the file to read",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path := GetString(params, "path")
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var sb strings.Builder
	for i, line := range lines {
		fmt.Fprintf(&sb, "%4d | %s\n", i+1, line)
	}
	return sb.String(), nil
}
