package safety

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Level int

const (
	Safe Level = iota
	NeedsApproval
)

// Checker manages tool approval.
type Checker struct {
	autoApproveAll   bool
	sessionApprovals map[string]bool // tool name -> always approved
	reader           io.Reader
	writer           io.Writer
}

func NewChecker(autoApproveAll bool) *Checker {
	return &Checker{
		autoApproveAll:   autoApproveAll,
		sessionApprovals: make(map[string]bool),
		reader:           os.Stdin,
		writer:           os.Stderr,
	}
}

// Classify returns the safety level of a tool.
func Classify(toolName string) Level {
	switch toolName {
	case "ReadFile", "ListDirectory", "GetFileInfo", "SearchFiles":
		return Safe
	default:
		return NeedsApproval
	}
}

// Check returns true if the tool call is approved.
func (c *Checker) Check(toolName string, params map[string]any) (bool, error) {
	if c.autoApproveAll {
		return true, nil
	}

	level := Classify(toolName)
	if level == Safe {
		return true, nil
	}

	if c.sessionApprovals[toolName] {
		return true, nil
	}

	return c.promptUser(toolName, params)
}

func (c *Checker) promptUser(toolName string, params map[string]any) (bool, error) {
	// Show what the tool wants to do
	fmt.Fprintf(c.writer, "\n🔒 Tool %s wants to execute:\n", toolName)
	for k, v := range params {
		val := fmt.Sprintf("%v", v)
		if len(val) > 200 {
			val = val[:200] + "..."
		}
		fmt.Fprintf(c.writer, "  %s: %s\n", k, val)
	}
	fmt.Fprintf(c.writer, "\nApprove? [y/n/always] ")

	scanner := bufio.NewScanner(c.reader)
	if !scanner.Scan() {
		return false, fmt.Errorf("failed to read input")
	}

	response := strings.TrimSpace(strings.ToLower(scanner.Text()))
	switch response {
	case "y", "yes":
		return true, nil
	case "always", "a":
		c.sessionApprovals[toolName] = true
		return true, nil
	default:
		return false, nil
	}
}
