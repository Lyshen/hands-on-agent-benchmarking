package main

import (
	"errors"
	"os/exec"
)

type OpenAIAdapter struct {
}

func (adapter OpenAIAdapter) buildOpenAITool(tool Tool) OpenAITool {
	openAITool := OpenAITool{
		Type: "function",
		Function: Function{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters: Parameters{
				Type: "object",
				Properties: Properties{
					Command: Command{
						Type:        "string",
						Description: "The command to run.",
					},
				},
				Required: []string{
					"command",
				},
				AdditionalProperties: false,
			},
		},
	}

	return openAITool
}

func (adapter OpenAIAdapter) buildOpenAIRequest(contexter Contexter) OpenAIChatRequest {
	openAITools := []OpenAITool{}
	for _, tool := range contexter.Tools {
		openAITools = append(openAITools, adapter.buildOpenAITool(tool))
	}

	chatReq := OpenAIChatRequest{
		Model:    contexter.Model,
		Messages: contexter.Messages,
		Tools:    openAITools,
	}

	return chatReq
}

func (adapter OpenAIAdapter) buildMessage(response OpenAIChatResponse) Message {
	return response.Choices[0].Message
}

func (adapter OpenAIAdapter) buildToolInvocation(toolCall OpenAIToolCall) ToolInvocation {
	toolInvocation := ToolInvocation{
		ID:        toolCall.Id,
		Name:      toolCall.Function.Name,
		Arguments: toolCall.Function.Arguments,
	}

	return toolInvocation
}

func (adapter OpenAIAdapter) buildOpenAIToolResultMessage(toolResult ToolResult) Message {
	content := toolResult.Output
	if toolResult.Error != nil {
		var exitErr *exec.ExitError
		if errors.As(toolResult.Error, &exitErr) {
			content = "Command err " + string(exitErr.Stderr)
		} else {
			content = toolResult.Error.Error()
		}
	}

	message := Message{
		Role:       "tool",
		Content:    content,
		ToolCallID: toolResult.InvocationID,
	}
	return message
}
