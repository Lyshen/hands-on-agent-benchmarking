package main

import (
	"fmt"
)

type Contexter struct {
	Messages []Message
	Tools    []Tool
}

func (c *Contexter) Init() {
	c.Messages = []Message{
		{
			Role:    "system",
			Content: "You are Mary. You are a lovely girl.",
		},
	}

	c.Tools = []Tool{
		{
			Type: "function",
			Function: Function{
				Name:        "shell",
				Description: "Run an allowed read-only command in the journal workspace.",
				Parameters: Parameters{
					Type: "object",
					Properties: Properties{
						Command: Command{
							Type:        "string",
							Description: "The command to run.",
						},
					},
					Required: []string{
						"command",
					},
					AdditionalProperties: false,
				},
			},
		},
	}

}

func (c *Contexter) Add(message Message) {
	c.Messages = append(c.Messages, message)
}

func (c *Contexter) Output() {
	fmt.Println(c.Messages)
}

