package main

import (
	"errors"
	"fmt"
	"strings"
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

	//verificationLoopState := NewContext(conf.State.Model, "You are a code review agent.Review the completed work against the original task.Inspect the actual files and run relevant checks when useful.Do not modify any files. Report concrete problems that should be fixed.If no problems are found, respond with 'NO_ISSUES'.")
	verificationLoopState := NewContext(conf.State.Model, "You are an adversarial verification agent.Your goal is to find concrete, reproducible failures in the Main Agent's completed work.Start from the Original Task, not the Main Agent's claims. Identify the required behavior, constraints, and the input/output contract of the final deliverable.Inspect the actual implementation and determine how the deliverable can be tested.Design and execute tests at multiple applicable levels, including:- **Basic functionality:** Expected behavior on normal inputs.- **Boundary conditions:** Empty inputs, minimum and maximum values, unusual formats, and other edge cases.- **Invalid inputs:** Malformed data, missing fields, incorrect types, and error handling.- **Integration:** Interactions between components, interfaces, files, and dependencies.- **Cross-scenario behavior:** Combinations of conditions, sequences of operations, and state transitions that individual tests may miss.Prioritize tests that are most likely to expose incorrect assumptions, incomplete implementations, or violations of the Original Task.Actively try to break the deliverable using valid, relevant, and increasingly challenging test cases.For every failure, report:1. The requirement or expected behavior being violated.2. The exact test input or reproduction steps.3. The expected result.4. The actual observed result.5. The relevant error output or other supporting evidence.Do not modify the implementation, existing tests, benchmark verifier, or acceptance criteria. You may create temporary test data or test scripts in isolated locations when necessary, but avoid modifying the deliverable or leaving persistent side effects.Distinguish confirmed failures from suspected problems. Do not report hypothetical failures as verified defects.Focus only on the Original Task and its deliverable. Do not perform unrelated code reviews.If no failures are found, respond with exactly 'NO_ISSUES'.")
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

const MaxReviews = 15

func (r *Runner) Run(input UserItem) (AssistantItem, error) {
    result, err := r.mainLoop.runTurn(input)
    if err != nil {
        return AssistantItem{}, err
    }

    for i := 0; i < MaxReviews; i++ {
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

        // 最后一次验证仍失败，不再进行未经验证的修复
        if i == MaxReviews-1 {
            return result, fmt.Errorf(
                "verification did not pass after %d reviews: %s",
                MaxReviews,
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

/*
func (r *Runner) Run(input UserItem) (AssistantItem, error) {
    firstResult, err := r.mainLoop.runTurn(input)
	if err != nil {
        return AssistantItem{}, err
	}

	//return firstResult, nil
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
	return r.mainLoop.runTurn(repairInput)
}*/
