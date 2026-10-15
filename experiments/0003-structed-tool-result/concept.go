package main

type Tool struct {
	Name        string
	Description string
}

type Item struct {
	Kind             string
	SystemPromptItem *SystemPromptItem
	UserItem         *UserItem
	AssistantItem    *AssistantItem
	ToolInvocation   *ToolInvocation
	ToolResult       *ToolResult
}

type SystemPromptItem struct {
	Content string
}

type UserItem struct {
	Content string
}

type AssistantItem struct {
	Content string
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
