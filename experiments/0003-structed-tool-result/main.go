package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func runTurn(state *Context, provider *LLMProvider, adapter *OpenAIAdapter) (AssistantItem, error) {
	for {
		req := adapter.buildOpenAIRequest(*state)
		resp, err := provider.send(req)
		if err != nil {
			return AssistantItem{}, err
		}

		item, err := adapter.buildTranscriptItem(resp)
		if err != nil {
			return AssistantItem{}, err
		}

		switch value := item.(type) {
		case AssistantItem:
			state.Add(value)
			return value, nil

		case ToolCallsItem:
			state.Add(value)
			for _, toolInvocation := range value.Calls {
				toolResult := execute(toolInvocation)
				state.Add(ToolResultItem{Result: toolResult})
			}
		}
	}
}

func main() {
	conf, err := loadConfig("config.local.json")
	if err != nil {
		fmt.Println("fail to load config: ", err)
		return
	}

	state := Context{}
	state.Init(conf.Model)
	provider := LLMProvider{}
	provider.Init(&conf)
	adapter := OpenAIAdapter{}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n\nYou: ")

		userMsg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("fail to read user message ", err)
			continue
		}
		userMsg = strings.Trim(userMsg, "\n")

		state.Add(UserItem{Content: userMsg})

		assistantItem, err := runTurn(&state, &provider, &adapter)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Assistant: ", assistantItem.Content)
		}
	}
}
