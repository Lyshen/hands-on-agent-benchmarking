package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
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

func main() {
	conf, err := loadConfig("config.local.json")
	if err != nil {
		fmt.Println("fail to load config: ", err)
		return
	}

	messages := []Message{
		{
			Role:    "system",
			Content: "You are Mary. You are a lovely girl.",
		},
	}

	tools := initTools()
	reader := bufio.NewReader(os.Stdin)
	client := &http.Client{}
	for {
		fmt.Println(messages)
		fmt.Println("____________________")
		fmt.Print("You: ")

		userMsg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("fail to read user message ", err)
			continue
		}
		userMsg = strings.Trim(userMsg, "\n")

		messages = append(messages, Message{Role: "user", Content: userMsg})

		for {
			respMsg, err := triggeredByMessages(client, conf, messages, *tools)
			if err != nil {
				fmt.Println(err)
				break
			}

			messages = append(messages, *respMsg)

			for _, toolCall := range respMsg.ToolCalls {
				fmt.Println("ToolCall: ", toolCall)
				resultMsg := execute(toolCall)
				fmt.Println("ToolCallResult: ", resultMsg)
				messages = append(messages, resultMsg)
			}

			if len(respMsg.ToolCalls) == 0 {
				fmt.Println("Assistant: ", respMsg.Content)
				break
			}
		}
	}
}
