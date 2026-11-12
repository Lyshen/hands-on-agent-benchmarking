package main

import (
	"errors"
	"fmt"
	"strings"
)

type Loop struct {
	state    Context
	provider Provider
	conf     LoopConfig
}

func NewLoop(state Context, provider Provider, conf LoopConfig) Loop {
	return Loop{state: state, provider: provider, conf: conf}
}

func (loop *Loop) runTurn(input UserItem) (AssistantItem, error) {
	loop.state.Add(input)

	for step := 0; ; step++ {
		if loop.conf.MaxStepsPerTurn != nil && step >= *loop.conf.MaxStepsPerTurn {
			return AssistantItem{}, fmt.Errorf("max steps per turn exceeded: %d", *loop.conf.MaxStepsPerTurn)
		}

		if loop.conf.TokenBudget != nil && loop.state.TokenUsage.TotalToken > *loop.conf.TokenBudget {
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
				toolResult := execute(toolInvocation, loop.conf.ShellTimeoutSeconds)
				fmt.Printf(
					"ToolResult: output=%q error=%v\n",
					toolResult.Output,
					toolResult.Error,
				)
				loop.state.Add(ToolResultItem{Result: toolResult})
			}

		default:
			return AssistantItem{}, fmt.Errorf(
				"unsupported provider response type: %T",
				resp.Item,
			)
		}
	}
}

type Runner struct {
	maxReviews       int
	mainLoop         Loop
	verificationLoop Loop
}

func NewRunner(model string, provider Provider, conf AgentRunConfig) Runner {
	mainLoopState := NewContext(model, conf.MainLoop.SystemPrompt)
	mainLoop := NewLoop(
		mainLoopState,
		provider,
		conf.MainLoop,
	)

	verificationLoopState := NewContext(model, conf.VerificationLoop.SystemPrompt)
	verificationLoop := NewLoop(
		verificationLoopState,
		provider,
		conf.VerificationLoop,
	)

	return Runner{
		maxReviews:       conf.MaxReviews,
		mainLoop:         mainLoop,
		verificationLoop: verificationLoop,
	}
}

func (r *Runner) Run(input UserItem) (AssistantItem, error) {
	result, err := r.mainLoop.runTurn(input)
	if err != nil {
		return AssistantItem{}, err
	}

	for i := 0; i < r.maxReviews; i++ {
		reviewInput := UserItem{
			Content: fmt.Sprintf(
				"Original task:\n%s\n\nMain agent summary:\n%s\n\n"+
					"Verify the current workspace against the original task.",
				input.Content,
				result.Content,
			),
		}

		review, err := r.verificationLoop.runTurn(reviewInput)
		if err != nil {
			return AssistantItem{}, err
		}

		fmt.Printf("Verification %d: %s\n", i+1, review.Content)

		if strings.TrimSpace(review.Content) == "NO_ISSUES" {
			return result, nil
		}

		if i == r.maxReviews-1 {
			return result, fmt.Errorf(
				"verification did not pass after %d reviews: %s",
				r.maxReviews,
				review.Content,
			)
		}

		repairInput := UserItem{
			Content: fmt.Sprintf(
				"Verification found the following problems:\n%s\n\n"+
					"Fix the confirmed issues and complete the original task.",
				review.Content,
			),
		}

		result, err = r.mainLoop.runTurn(repairInput)
		if err != nil {
			return AssistantItem{}, err
		}
	}

	return result, nil
}
