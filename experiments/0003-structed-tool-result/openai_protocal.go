package main

type OpenAIChatRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
	Tools    []OpenAITool    `json:"tools"`
}

type OpenAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type OpenAIToolCall struct {
	Function OpenAIToolCallSchema `json:"function"`
	Id       string               `json:"id"`
	Type     string               `json:"type"`
}

type OpenAIToolCallSchema struct {
	Arguments string `json:"arguments"`
	Name      string `json:"name"`
}

type OpenAITool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type OpenAIChatResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message OpenAIMessage `json:"message"`
}

type ShellCommand struct {
	Command string `json:"command"`
}
