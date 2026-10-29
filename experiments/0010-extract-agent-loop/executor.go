package main

import (
	"encoding/json"
	"errors"
	"os/exec"
)

func execute(toolInvocation ToolInvocation) ToolResult {
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

	cmd := exec.Command("bash", "-o", "pipefail", "-c", shellCommand.Command)
	output, err := cmd.Output()

	var exitErr *exec.ExitError
	if err == nil {
		result.Output = string(output)
	} else if errors.As(err, &exitErr) {
		result.Error = errors.New("Command err " + string(exitErr.Stderr))
	} else {
		result.Error = err
	}

	return result
}
