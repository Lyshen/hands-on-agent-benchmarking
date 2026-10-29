package main

type OpenAIChatRequest struct {
	Model           string          `json:"model"`
	Messages        []OpenAIMessage `json:"messages"`
	Tools           []OpenAITool    `json:"tools"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
}

type OpenAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content"`
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
	Type     string         `json:"type"`
	Function OpenAIFunction `json:"function"`
}

type OpenAIFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type OpenAIChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIChatResponse struct {
	Choices []Choice        `json:"choices"`
	Usage   OpenAIChatUsage `json:"usage"`
}

type Choice struct {
	Message OpenAIMessage `json:"message"`
}

type ShellCommand struct {
	Command string `json:"command"`
}
