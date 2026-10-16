package main

import (
	"errors"
)

type OpenAIAdapter struct {
}

func (adapter OpenAIAdapter) buildOpenAITool(tool Tool) OpenAITool {
	return OpenAITool{
		Type: "function",
		Function: Function{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "The command to run.",
					},
				},
				"required": []string{
					"command",
				},
				"additionalProperties": false,
			},
		},
	}
}

func (adapter OpenAIAdapter) buildOpenAIMessage(item TranscriptItem) OpenAIMessage {
	var role string
	var content string
	switch value := item.(type) {
	case SystemPromptItem:
		role = "system"
		content = value.Content
	case UserItem:
		role = "user"
		content = value.Content
	case AssistantItem:
		role = "assistant"
		content = value.Content
	case ToolCallsItem:
		return adapter.buildOpenAIToolCallMessage(value)
	case ToolResultItem:
		return adapter.buildOpenAIToolResultMessage(value)
	}
	return OpenAIMessage{
		Role:    role,
		Content: content,
	}
}

func (adapter OpenAIAdapter) buildOpenAIRequest(context Context) OpenAIChatRequest {
	openAITools := []OpenAITool{}
	for _, tool := range context.Tools {
		openAITools = append(openAITools, adapter.buildOpenAITool(tool))
	}

	openAIMessages := []OpenAIMessage{}
	for _, item := range context.Transcript {
		openAIMessages = append(openAIMessages, adapter.buildOpenAIMessage(item))
	}

	chatReq := OpenAIChatRequest{
		Model:    context.Model,
		Messages: openAIMessages,
		Tools:    openAITools,
	}

	return chatReq
}

func (adapter OpenAIAdapter) buildTranscriptItem(response OpenAIChatResponse) (TranscriptItem, error) {
	if len(response.Choices) == 0 {
		return AssistantItem{}, errors.New("no message found in response choices")
	}

	if len(response.Choices[0].Message.ToolCalls) == 0 {
		return AssistantItem{Content: response.Choices[0].Message.Content}, nil
	}

	toolInvocations := []ToolInvocation{}
	for _, toolCall := range response.Choices[0].Message.ToolCalls {
		toolInvocations = append(toolInvocations, ToolInvocation{
			ID:        toolCall.Id,
			Name:      toolCall.Function.Name,
			Arguments: toolCall.Function.Arguments,
		})
	}
	toolCallsItem := ToolCallsItem{
		Calls: toolInvocations,
	}

	return toolCallsItem, nil
}

func (adapter OpenAIAdapter) buildOpenAIToolCallMessage(item ToolCallsItem) OpenAIMessage {
	calls := []OpenAIToolCall{}
	for _, call := range item.Calls {
		calls = append(calls, OpenAIToolCall{
			Id: call.ID,
			Function: OpenAIToolCallSchema{
				Name:      call.Name,
				Arguments: call.Arguments,
			},
			Type: "function",
		})
	}

	message := OpenAIMessage{
		Role:      "assistant",
		ToolCalls: calls,
	}

	return message
}

func (adapter OpenAIAdapter) buildOpenAIToolResultMessage(item ToolResultItem) OpenAIMessage {
	content := item.Result.Output
	if item.Result.Error != nil {
		content = item.Result.Error.Error()
	}

	message := OpenAIMessage{
		Role:       "tool",
		Content:    content,
		ToolCallID: item.Result.InvocationID,
	}
	return message
}
