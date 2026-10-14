package main

type Message struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type OpenAIChatRequest struct {
	Model    string       `json:"model"`
	Messages []Message    `json:"messages"`
	Tools    []OpenAITool `json:"tools"`
}

type Choice struct {
	Message Message `json:"message"`
}

type OpenAIChatResponse struct {
	Choices []Choice `json:"choices"`
}

// Tool

type ShellCommand struct {
	Command string `json:"command"`
}

type OpenAIToolCallSchema struct {
	Arguments string `json:"arguments"`
	Name      string `json:"name"`
}

type OpenAIToolCall struct {
	Function OpenAIToolCallSchema `json:"function"`
	Id       string               `json:"id"`
	Type     string               `json:"type"`
}

type Command struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Properties struct {
	Command Command `json:"command"`
}

type Parameters struct {
	Type                 string     `json:"type"`
	Properties           Properties `json:"properties"`
	Required             []string   `json:"required"`
	AdditionalProperties bool       `json:"additionalProperties"`
}

type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

type OpenAITool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}
