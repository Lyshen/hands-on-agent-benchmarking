package main

type OllamaChatRequest struct {
	Model    string          `json:"model"`
	Think    bool            `json:"think"`
	Stream   bool            `json:"stream"`
	Messages []OllamaMessage `json:"messages"`
	Tools    []OllamaTool    `json:"tools"`
}

type OllamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []OllamaToolCall `json:"tool_calls,omitempty"`
}

type OllamaToolCall struct {
	Id       string               `json:"id"`
	Function OllamaToolCallSchema `json:"function"`
}

type OllamaToolCallSchema struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type OllamaTool struct {
	Type     string         `json:"type"`
	Function OllamaFunction `json:"function"`
}

type OllamaFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type OllamaChatResponse struct {
	Message OllamaMessage `json:"message"`
	Done    bool          `json:"done"`
}
