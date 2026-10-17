package main

type Context struct {
	Model      string
	Transcript []TranscriptItem
	Tools      []Tool
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
}

func (c *Context) Add(item TranscriptItem) {
	c.Transcript = append(c.Transcript, item)
}
