package main

import (
	"errors"
	"encoding/json"
	"fmt"
)

type OllamaAdapter struct {
}

func (adapter OllamaAdapter) buildTool(tool Tool) (OllamaTool, error) {
	properties := make(map[string]any, len(tool.Parameters))
	required := make([]string, 0, len(tool.Parameters))

	for _, parameter := range tool.Parameters {
		var schemaType string

		switch parameter.Type {
		case ParameterString:
			schemaType = "string"
		default:
			return OllamaTool{}, errors.New(tool.Name + " unsupported parameter type ")
		}

		properties[parameter.Name] = map[string]any{
			"type":        schemaType,
			"description": parameter.Description,
		}

		if parameter.Required {
			required = append(required, parameter.Name)
		}
	}

	return OllamaTool{
		Type: "function",
		Function: OllamaFunction{
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

func (adapter OllamaAdapter) buildMessage(item TranscriptItem) (OllamaMessage, error) {
	switch value := item.(type) {
	case SystemPromptItem:
		return OllamaMessage{Role: "system", Content: value.Content}, nil
	case UserItem:
		return OllamaMessage{Role: "user", Content: value.Content}, nil
	case AssistantItem:
		return OllamaMessage{Role: "assistant", Content: value.Content}, nil
	case ToolCallsItem:
		return adapter.buildToolCallMessage(value)
	case ToolResultItem:
		return adapter.buildToolResultMessage(value), nil
	}
	return OllamaMessage{}, errors.New("TranscriptItem type is not found, please check")
}

func (adapter OllamaAdapter) buildRequest(context Context) (OllamaChatRequest, error) {
	ollamaTools := []OllamaTool{}
	for _, tool := range context.Tools {
		ollamaTool, err := adapter.buildTool(tool)
		if err == nil {
			ollamaTools = append(ollamaTools, ollamaTool)
		} else {
			return OllamaChatRequest{}, err
		}
	}

	messages := []OllamaMessage{}
	for _, item := range context.Transcript {
		message, err := adapter.buildMessage(item)
		if err != nil {
			return OllamaChatRequest{}, err
		}
		messages = append(messages, message)
	}

	chatReq := OllamaChatRequest{
		Model:    context.Model,
		Messages: messages,
		Tools:    ollamaTools,
	}

	return chatReq, nil
}

func (adapter OllamaAdapter) buildTranscriptItem(response OllamaChatResponse) (TranscriptItem, error) {
    fmt.Println(response.Message)
    fmt.Println(response.Message.ToolCalls)
	if len(response.Message.ToolCalls) == 0 {
		return AssistantItem{Content: response.Message.Content}, nil
	}
    fmt.Println(len(response.Message.ToolCalls))
	toolInvocations := []ToolInvocation{}
	for _, toolCall := range response.Message.ToolCalls {
        data, err := json.Marshal(toolCall.Function.Arguments)
		if err != nil {
            return AssistantItem{}, err
		}

		toolInvocations = append(toolInvocations, ToolInvocation{
			ID: toolCall.Id,
			Name:      toolCall.Function.Name,
			Arguments: string(data),
		})
	}
	toolCallsItem := ToolCallsItem{
		Calls: toolInvocations,
	}

	return toolCallsItem, nil
}

func (adapter OllamaAdapter) buildToolCallMessage(item ToolCallsItem) (OllamaMessage, error) {
	calls := []OllamaToolCall{}
	for _, call := range item.Calls {
		arguments := map[string]any{}
		err := json.Unmarshal([]byte(call.Arguments), &arguments)
		if err != nil {
            return OllamaMessage{}, err
		}

		calls = append(calls, OllamaToolCall{
			Function: OllamaToolCallSchema{
				Name:      call.Name,
				Arguments: arguments,
			},
		})
	}

	message := OllamaMessage{
		Role:      "assistant",
		ToolCalls: calls,
	}

	return message, nil
}

func (adapter OllamaAdapter) buildToolResultMessage(item ToolResultItem) OllamaMessage {
	content := item.Result.Output
	if item.Result.Error != nil {
		content = item.Result.Error.Error()
	}

	message := OllamaMessage{
		Role:       "tool",
		Content:    content,
	}
	return message
}
