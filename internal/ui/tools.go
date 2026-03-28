package ui

import (
	"fmt"
	"strings"
)

// FormatToolCall formats a tool call for display, showing the tool name and
// the most relevant parameters.
func FormatToolCall(name string, params map[string]any) string {
	parts := formatToolParams(name, params)
	if len(parts) == 0 {
		return fmt.Sprintf("🔧 %s", name)
	}
	return fmt.Sprintf("🔧 %s(%s)", name, strings.Join(parts, ", "))
}

func formatToolParams(name string, params map[string]any) []string {
	var parts []string
	switch name {
	case "ReadFile", "WriteFile", "EditFile", "ListDirectory", "GetFileInfo":
		if path, ok := params["path"]; ok {
			parts = append(parts, fmt.Sprintf("path: %q", path))
		}
	case "SearchFiles":
		if pattern, ok := params["pattern"]; ok {
			parts = append(parts, fmt.Sprintf("pattern: %q", pattern))
		}
		if path, ok := params["path"]; ok {
			parts = append(parts, fmt.Sprintf("path: %q", path))
		}
	case "RunCommand":
		if cmd, ok := params["command"]; ok {
			cmdStr := fmt.Sprintf("%v", cmd)
			if len(cmdStr) > 60 {
				cmdStr = cmdStr[:60] + "..."
			}
			parts = append(parts, fmt.Sprintf("command: %q", cmdStr))
		}
	default:
		if path, ok := params["path"]; ok {
			parts = append(parts, fmt.Sprintf("path: %q", path))
		}
	}
	return parts
}
