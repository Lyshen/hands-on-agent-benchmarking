package main

import (
	"errors"
	"fmt"
)

type Loop struct {
	state       *Context
	provider    Provider
	tokenBudget *int
}

type LoopConfig struct {
	State       *Context
	Provider    Provider
	TokenBudget *int
}

func NewLoop(state *Context, provider Provider, tokenBudget *int) Loop {
	return Loop{
		state:       state,
		provider:    provider,
		tokenBudget: tokenBudget,
	}
}

func (loop Loop) runTurn(input UserItem) (AssistantItem, error) {
	loop.state.Add(input)

	for {
		if loop.tokenBudget != nil && loop.state.TokenUsage.TotalToken > *loop.tokenBudget {
			return AssistantItem{}, errors.New("current loop is out of token budget")
		}

		resp, err := loop.provider.Generate(loop.state)
		if err != nil {
			return AssistantItem{}, err
		}

		switch value := resp.Item.(type) {
		case AssistantItem:
			loop.state.Add(value)
			return value, nil

		case ToolCallsItem:
			loop.state.Add(value)
			for _, toolInvocation := range value.Calls {
				fmt.Printf(
					"ToolCall: name=%s arguments=%s\n",
					toolInvocation.Name,
					toolInvocation.Arguments,
				)
				toolResult := execute(toolInvocation)
				fmt.Printf(
					"ToolResult: output=%q error=%v\n",
					toolResult.Output,
					toolResult.Error,
				)
				loop.state.Add(ToolResultItem{Result: toolResult})
			}
		}
	}
}

type Runner struct {
	mainLoop Loop
}

func NewRunner(conf LoopConfig) Runner {
	return Runner{
		mainLoop: NewLoop(
			conf.State,
			conf.Provider,
			conf.TokenBudget,
		),
	}
}

func (r Runner) Run(input UserItem) (AssistantItem, error) {
	return r.mainLoop.runTurn(input)
}
