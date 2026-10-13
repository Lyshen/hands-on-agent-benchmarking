package main

import (
	"encoding/json"
	"errors"
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

func (adapter OpenAIAdapter) buildRequest(contexter Contexter) ([]byte, error) {
	openAITools := []OpenAITool{}
	for _, tool := range contexter.Tools {
		openAITools = append(openAITools, adapter.buildOpenAITool(tool))
	}

	chatReq := ChatRequest{
		Model:    contexter.Model,
		Messages: contexter.Messages,
		Tools:    openAITools,
	}

	var reqBytes []byte
	reqBytes, err := json.Marshal(chatReq)
	if err != nil {
		return nil, err
	}

	return reqBytes, nil
}

func (adapter OpenAIAdapter) parseResponse(respBytes []byte) (Message, error) {
	var resp ChatResponse
	msg := Message{}
	err := json.Unmarshal(respBytes, &resp)
	if err != nil {
		return msg, err
	}

	if len(resp.Choices) == 0 {
		return msg, errors.New("no message found in response choices")
	}

	return resp.Choices[0].Message, nil
}

func (adapter OpenAIAdapter) buildToolInvocation(toolCall OpenAIToolCall) ToolInvocation {
	toolInvocation := ToolInvocation{
		Id:        toolCall.Id,
		Name:      toolCall.Function.Name,
		Arguments: toolCall.Function.Arguments,
	}

	return toolInvocation
}
