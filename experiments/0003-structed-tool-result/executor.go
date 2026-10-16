package main

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

func execute(toolInvocation ToolInvocation) ToolResult {
	result := ToolResult{
		InvocationID: toolInvocation.ID,
		Output:       "",
		Error:        nil,
	}

	if toolInvocation.Name == "shell" {
		var shellCommand ShellCommand
		err := json.Unmarshal([]byte(toolInvocation.Arguments), &shellCommand)

		if err == nil {
			fields := strings.Fields(shellCommand.Command)

			if len(fields) >= 1 {
				name := fields[0]
				args := fields[1:]

				if name == "ls" {
					cmd := exec.Command(name, args...)
					output, err := cmd.Output()

					var exitErr *exec.ExitError
					if err == nil {
						result.Output = string(output)
					} else if errors.As(err, &exitErr) {
						result.Error = errors.New("Command err " + string(exitErr.Stderr))
					} else {
						result.Error = err
					}
				} else {
					result.Error = errors.New("ToolCall Error: command is not in approval list. Only ls is allowed.")
				}
			} else {
				result.Error = errors.New("ToolCall Error: shell command is empty. Please Check.")
			}
		} else {
			result.Error = errors.New("ToolCall Error: shell command needs to be JSON")
		}
	} else {
		result.Error = errors.New("ToolCall Error: toolcall name supports 'shell' only.")
	}

	return result
}
