package main

type TokenUsage struct {
	InputToken  int
	OutputToken int
	TotalToken  int
}

type Context struct {
	Model      string
	Transcript []TranscriptItem
	Tools      []Tool
	TokenUsage TokenUsage
}

func (c *Context) Init(model string) {
	c.Model = model

	c.Transcript = []TranscriptItem{
		SystemPromptItem{
			Content: "You are Mary. You are a lovely girl.",
		},
	}

	c.Tools = []Tool{
		{
			Name:        "shell",
			Description: "Run an allowed read-only command in the journal workspace.",
			Parameters: []ToolParameter{
				{
					Name:        "command",
					Description: "The command to run.",
					Type:        ParameterString,
					Required:    true,
				},
			},
		},
	}

	c.TokenUsage = TokenUsage{}
}

func (c *Context) Add(item TranscriptItem) {
	c.Transcript = append(c.Transcript, item)
}

func (c *Context) AddUsage(usage TokenUsage) {
	c.TokenUsage.InputToken += usage.InputToken
	c.TokenUsage.OutputToken += usage.OutputToken
	c.TokenUsage.TotalToken += usage.TotalToken
}
