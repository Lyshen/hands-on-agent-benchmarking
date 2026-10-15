package main

import (
	"errors"
	"os/exec"
)

type OpenAIAdapter struct {
}

func (adapter OpenAIAdapter) buildOpenAITool(tool Tool) OpenAITool {
	return OpenAITool{
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
}

func (adapter OpenAIAdapter) buildOpenAIMessage(item Item) OpenAIMessage {
	var role string
	var content string
	switch item.Kind {
	case "system":
		role = "system"
		content = item.SystemPromptItem.Content
	case "user":
		role = "user"
		content = item.UserItem.Content
	case "assistant":
		role = "assistant"
		content = item.AssistantItem.Content
	case "toolInvocation":
		return adapter.buildOpenAIToolCallMessage(*item.ToolInvocation)
	case "toolResult":
		return adapter.buildOpenAIToolResultMessage(*item.ToolResult)
	}
	return OpenAIMessage{
		Role:    role,
		Content: content,
	}
}

func (adapter OpenAIAdapter) buildOpenAIRequest(contexter Contexter) OpenAIChatRequest {
	openAITools := []OpenAITool{}
	for _, tool := range contexter.Tools {
		openAITools = append(openAITools, adapter.buildOpenAITool(tool))
	}

	openAIMessages := []OpenAIMessage{}
	for _, item := range contexter.Items {
		openAIMessages = append(openAIMessages, adapter.buildOpenAIMessage(item))
	}

	chatReq := OpenAIChatRequest{
		Model:    contexter.Model,
		Messages: openAIMessages,
		Tools:    openAITools,
	}

	return chatReq
}

func (adapter OpenAIAdapter) buildMessage(response OpenAIChatResponse) OpenAIMessage {
	return response.Choices[0].Message
}

func (adapter OpenAIAdapter) buildToolInvocation(toolCall OpenAIToolCall) ToolInvocation {
	return ToolInvocation{
		ID:        toolCall.Id,
		Name:      toolCall.Function.Name,
		Arguments: toolCall.Function.Arguments,
	}
}

func (adapter OpenAIAdapter) buildOpenAIToolCallMessage(toolInvocation ToolInvocation) OpenAIMessage {
	message := OpenAIMessage{
		Role: "assistant",
		ToolCalls: []OpenAIToolCall{
			{
				Id: toolInvocation.ID,
				Function: OpenAIToolCallSchema{
					Name:      toolInvocation.Name,
					Arguments: toolInvocation.Arguments,
				},
				Type: "function",
			},
		},
	}

	return message
}

func (adapter OpenAIAdapter) buildOpenAIToolResultMessage(toolResult ToolResult) OpenAIMessage {
	content := toolResult.Output
	if toolResult.Error != nil {
		var exitErr *exec.ExitError
		if errors.As(toolResult.Error, &exitErr) {
			content = "Command err " + string(exitErr.Stderr)
		} else {
			content = toolResult.Error.Error()
		}
	}

	message := OpenAIMessage{
		Role:       "tool",
		Content:    content,
		ToolCallID: toolResult.InvocationID,
	}
	return message
}
