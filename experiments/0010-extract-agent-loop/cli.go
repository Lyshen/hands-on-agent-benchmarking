package main

import (
	"fmt"
)

func runCli(task string) {
	conf, err := loadConfigFromEnv()
	if err != nil {
		fmt.Println("fail to load config: ", err)
		return
	}

	provider, err := NewProvider(&conf.LLMProviderConfig)
	if err != nil {
		fmt.Println("fail to init provider: ", err)
		return
	}

	state := NewContext(conf.LLMProviderConfig.Model)
	runner := NewRunner(LoopConfig{
		State:       &state,
		Provider:    provider,
		TokenBudget: conf.AgentRunConfig.TokenBudget,
	})

	assistantItem, err := runner.Run(UserItem{Content: task})

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Assistant: ", assistantItem.Content)
	}
}
