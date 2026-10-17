package main

import (
	"errors"
)

type OpenAIAdapter struct {
}

func (adapter OpenAIAdapter) buildOpenAITool(tool Tool) (OpenAITool, error) {
	properties := make(map[string]any, len(tool.Parameters))
	required := make([]string, 0, len(tool.Parameters))

	for _, parameter := range tool.Parameters {
		var schemaType string

		switch parameter.Type {
		case ParameterString:
			schemaType = "string"
		default:
			return OpenAITool{}, errors.New(tool.Name + " unsupported parameter type ")
		}

		properties[parameter.Name] = map[string]any{
			"type":        schemaType,
			"description": parameter.Description,
		}

		if parameter.Required {
			required = append(required, parameter.Name)
		}
	}

	return OpenAITool{
		Type: "function",
		Function: Function{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters: map[string]any{
				"type":                 "object",
				"properties":           properties,
				"required":             required,
				"additionalProperties": false,
			},
		},
	}, nil
}

func (adapter OpenAIAdapter) buildOpenAIMessage(item TranscriptItem) (OpenAIMessage, error) {
	switch value := item.(type) {
	case SystemPromptItem:
		return OpenAIMessage{Role: "system", Content: value.Content}, nil
	case UserItem:
		return OpenAIMessage{Role: "user", Content: value.Content}, nil
	case AssistantItem:
		return OpenAIMessage{Role: "assistant", Content: value.Content}, nil
	case ToolCallsItem:
		return adapter.buildOpenAIToolCallMessage(value), nil
	case ToolResultItem:
		return adapter.buildOpenAIToolResultMessage(value), nil
	}
	return OpenAIMessage{}, errors.New("TranscriptItem type is not found, please check")
}

func (adapter OpenAIAdapter) buildOpenAIRequest(context Context) (OpenAIChatRequest, error) {
	openAITools := []OpenAITool{}
	for _, tool := range context.Tools {
		openAITool, err := adapter.buildOpenAITool(tool)
		if err == nil {
			openAITools = append(openAITools, openAITool)
		} else {
			return OpenAIChatRequest{}, err
		}
	}

	openAIMessages := []OpenAIMessage{}
	for _, item := range context.Transcript {
		message, err := adapter.buildOpenAIMessage(item)
		if err != nil {
			return OpenAIChatRequest{}, err
		}
		openAIMessages = append(openAIMessages, message)
	}

	chatReq := OpenAIChatRequest{
		Model:    context.Model,
		Messages: openAIMessages,
		Tools:    openAITools,
	}

	return chatReq, nil
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
