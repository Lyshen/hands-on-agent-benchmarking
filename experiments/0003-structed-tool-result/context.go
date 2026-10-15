package main

import (
	"fmt"
)

type Contexter struct {
	Model string
	Items []Item
	Tools []Tool
}

func (c *Contexter) Init(model string) {
	c.Model = model

	c.Items = []Item{
		{
			Kind:             "system",
			SystemPromptItem: &SystemPromptItem{Content: "You are Mary. You are a lovely girl."},
		},
	}

	c.Tools = []Tool{
		{
			Name:        "shell",
			Description: "Run an allowed read-only command in the journal workspace.",
		},
	}
}

func (c *Contexter) Add(item Item) {
	c.Items = append(c.Items, item)
}

func (c *Contexter) Output() {
	for _, item := range c.Items {
		fmt.Println(item.Kind)
	}
}
