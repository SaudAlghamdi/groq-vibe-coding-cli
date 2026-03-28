package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type RunCommandTool struct{}

func (t *RunCommandTool) Name() string { return "RunCommand" }
func (t *RunCommandTool) Description() string {
	return "Execute a shell command and return its output. The command runs in the current working directory with a 30-second timeout."
}

func (t *RunCommandTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The shell command to execute",
			},
		},
		"required": []string{"command"},
	}
}

func (t *RunCommandTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	command := GetString(params, "command")
	if command == "" {
		return "", fmt.Errorf("command is required")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var result string
	if stdout.Len() > 0 {
		result = stdout.String()
	}
	if stderr.Len() > 0 {
		if result != "" {
			result += "\n"
		}
		result += "STDERR:\n" + stderr.String()
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return result, fmt.Errorf("command timed out after 30 seconds")
		}
		if result == "" {
			return "", fmt.Errorf("command failed: %w", err)
		}
		// Return output even on non-zero exit
		return result + "\n(exit code: " + err.Error() + ")", nil
	}

	if result == "" {
		result = "(no output)"
	}
	return result, nil
}
