package main

import (
	"errors"
	"fmt"
)

type Loop struct {
	state       Context
	provider    Provider
	tokenBudget *int
}

type LoopConfig struct {
	State       Context
	Provider    Provider
	TokenBudget *int
}

func NewLoop(state Context, provider Provider, tokenBudget *int) Loop {
	return Loop{state: state, provider: provider, tokenBudget: tokenBudget}
}

func (loop *Loop) runTurn(input UserItem) (AssistantItem, error) {
	loop.state.Add(input)

	for {
		if loop.tokenBudget != nil && loop.state.TokenUsage.TotalToken > *loop.tokenBudget {
			return AssistantItem{}, errors.New("current loop is out of token budget")
		}

		resp, err := loop.provider.Generate(loop.state)
		if err != nil {
			return AssistantItem{}, err
		}

		loop.state.AddUsage(resp.Usage)

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
	mainLoop         Loop
	verificationLoop Loop
}

func NewRunner(conf LoopConfig) Runner {
	mainLoop := NewLoop(
		conf.State,
		conf.Provider,
		conf.TokenBudget,
	)

	verificationLoopState := NewContext(conf.State.Model, "You are a code review agent.Review the completed work against the original task.Inspect the actual files and run relevant checks when useful.Do not modify any files. Report concrete problems that should be fixed.If no problems are found, respond with 'NO_ISSUES'.")
	verificationLoop := NewLoop(
		verificationLoopState,
		conf.Provider,
		conf.TokenBudget,
	)

	return Runner{
		mainLoop:         mainLoop,
		verificationLoop: verificationLoop,
	}
}

func (r *Runner) Run(input UserItem) (AssistantItem, error) {
    firstResult, err := r.mainLoop.runTurn(input)
	if err != nil {
        return AssistantItem{}, err
	}

	return firstResult, nil
	/*
	reviewInput := UserItem{
        Content: fmt.Sprintf(
            "Original task:\n%s\n\nMain agent summary:\n%s\n\nReview the actual workspace.",
            input.Content,
            firstResult.Content,
        ),
    }

	reviewResult, err := r.verificationLoop.runTurn(reviewInput) 
	if err != nil {
        return AssistantItem{}, err
	}
	fmt.Println("reviewResult:" , reviewResult.Content)
    if strings.TrimSpace(reviewResult.Content) == "NO_ISSUES" {
        return firstResult, nil
    }

    repairInput := UserItem{
        Content: fmt.Sprintf(
            "Review feedback:\n%s\n\nFix these issues.",
            reviewResult.Content,
        ),
    }
	return r.mainLoop.runTurn(repairInput)*/
}
