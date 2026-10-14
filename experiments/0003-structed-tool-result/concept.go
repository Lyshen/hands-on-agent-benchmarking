package main

type Tool struct {
	Name        string
	Description string
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
