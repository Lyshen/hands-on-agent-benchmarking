package main

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"time"
)

func execute(toolInvocation ToolInvocation, timeout int) ToolResult {
	result := ToolResult{
		Invocation: toolInvocation,
		Output:     "",
		Error:      nil,
	}

	if toolInvocation.Name != "shell" {
		result.Error = errors.New("ToolCall Error: toolcall name supports 'shell' only.")
		return result
	}

	var shellCommand ShellCommand
	err := json.Unmarshal([]byte(toolInvocation.Arguments), &shellCommand)

	if err != nil {
		result.Error = errors.New("ToolCall Error: shell command needs to be JSON")
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-o", "pipefail", "-c", shellCommand.Command)
	output, err := cmd.CombinedOutput()

	result.Output = string(output)
	result.Error = err

	return result
}
