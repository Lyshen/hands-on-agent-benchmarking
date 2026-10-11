package main


type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

// Tool

type ShellCommand struct {
	Command string `json:"command"`
}

type ToolCallSchema struct {
	Arguments string `json:"arguments"`
	Name      string `json:"name"`
}

type ToolCall struct {
	Function ToolCallSchema `json:"function"`
	Id       string         `json:"id"`
	Type     string         `json:"type"`
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

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

