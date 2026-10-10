package main

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

func execute(toolCall ToolCall) Message {
	var message Message

	if toolCall.Function.Name == "shell" {
		var shellCommand ShellCommand
		err := json.Unmarshal([]byte(toolCall.Function.Arguments), &shellCommand)
		if err != nil {
			message = Message{
				Role:       "tool",
				ToolCallID: toolCall.Id,
				Content:    "Invalid arguments. Error in parsing. Please check",
			}
			return message
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
						message = Message{
							Role:       "tool",
							ToolCallID: toolCall.Id,
							Content:    "Command err " + string(exitErr.Stderr),
						}
					} else {
						message = Message{
							Role:       "tool",
							ToolCallID: toolCall.Id,
							Content:    "Command err " + err.Error(),
						}
					}
					return message
				}
				message = Message{
					Role:       "tool",
					ToolCallID: toolCall.Id,
					Content:    string(output),
				}
			} else {
				message = Message{
					Role:       "tool",
					ToolCallID: toolCall.Id,
					Content:    "Command is not in approval list. Only ls is allowed.",
				}
			}
		} else {
			message = Message{
				Role:       "tool",
				ToolCallID: toolCall.Id,
				Content:    "ToolCall command is empty. Please Check.",
			}

		}
	} else {
		message = Message{
			Role:       "tool",
			ToolCallID: toolCall.Id,
			Content:    "ToolCall Error: toolcall name is not found.",
		}
	}

	return message
}

