package main

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

func execute(toolInvocation ToolInvocation) ToolResult {
	result := ToolResult{
		Id:      toolInvocation.Id,
		Content: "",
	}

	if toolInvocation.Name == "shell" {
		var shellCommand ShellCommand
		err := json.Unmarshal([]byte(toolInvocation.Arguments), &shellCommand)
		if err != nil {
			result.Content = "Invalid arguments. Error in parsing. Please check"
			return result
		}

		fields := strings.Fields(shellCommand.Command)
		if len(fields) >= 1 {
			name := fields[0]
			args := fields[1:]

			if name == "ls" {
				cmd := exec.Command(name, args...)

				output, err := cmd.Output()
				if err != nil {
					var exitErr *exec.ExitError
					if errors.As(err, &exitErr) {
						result.Content = "Command err " + string(exitErr.Stderr)
					} else {
						result.Content = "Command err " + err.Error()
					}
				} else {
					result.Content = string(output)
				}
			} else {
				result.Content = "Command is not in approval list. Only ls is allowed."
			}
		} else {
			result.Content = "ToolCall command is empty. Please Check."
		}
	} else {
		result.Content = "ToolCall Error: toolcall name is not found."
	}

	return result
}
