package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GetFileInfoTool struct{}

func (t *GetFileInfoTool) Name() string { return "GetFileInfo" }
func (t *GetFileInfoTool) Description() string {
	return "Get metadata about a file including size, modification time, and detected language."
}

func (t *GetFileInfoTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path of the file to inspect",
			},
		},
		"required": []string{"path"},
	}
}

func (t *GetFileInfoTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path := GetString(params, "path")
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("getting file info: %w", err)
	}

	lang := detectLanguage(path)
	return fmt.Sprintf("File: %s\nSize: %s\nModified: %s\nIs Directory: %v\nLanguage: %s\nPermissions: %s",
		path,
		formatSize(info.Size()),
		info.ModTime().Format(time.RFC3339),
		info.IsDir(),
		lang,
		info.Mode().String(),
	), nil
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	languages := map[string]string{
		".go":    "Go",
		".py":    "Python",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".tsx":   "TypeScript (React)",
		".jsx":   "JavaScript (React)",
		".rs":    "Rust",
		".java":  "Java",
		".c":     "C",
		".cpp":   "C++",
		".h":     "C/C++ Header",
		".rb":    "Ruby",
		".php":   "PHP",
		".swift": "Swift",
		".kt":    "Kotlin",
		".sh":    "Shell",
		".yaml":  "YAML",
		".yml":   "YAML",
		".json":  "JSON",
		".xml":   "XML",
		".html":  "HTML",
		".css":   "CSS",
		".md":    "Markdown",
		".sql":   "SQL",
		".toml":  "TOML",
	}
	if lang, ok := languages[ext]; ok {
		return lang
	}
	return "Unknown"
}
