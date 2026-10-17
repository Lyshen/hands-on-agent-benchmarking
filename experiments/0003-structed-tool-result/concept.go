package main

type TranscriptItem interface {
	isTranscriptItem()
}

type SystemPromptItem struct {
	Content string
}

func (i SystemPromptItem) isTranscriptItem() {}

type UserItem struct {
	Content string
}

func (i UserItem) isTranscriptItem() {}

type AssistantItem struct {
	Content string
}

func (i AssistantItem) isTranscriptItem() {}

type ToolCallsItem struct {
	Calls []ToolInvocation
}

func (i ToolCallsItem) isTranscriptItem() {}

type ToolResultItem struct {
	Result ToolResult
}

func (i ToolResultItem) isTranscriptItem() {}

type ParameterType string

const (
	ParameterString ParameterType = "string"
)

type ToolParameter struct {
	Name        string
	Description string
	Type        ParameterType
	Required    bool
}

type Tool struct {
	Name        string
	Description string
	Parameters  []ToolParameter
}

type ToolInvocation struct {
	ID        string
	Name      string
	Arguments string
}

type ToolResult struct {
	InvocationID string
	Output       string
	Error        error
}
