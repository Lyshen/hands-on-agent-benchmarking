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
	loop := NewLoop(&state, provider, conf.AgentRunConfig.TokenBudget)

	state.Add(UserItem{Content: task})
	assistantItem, err := loop.runTurn()

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Assistant: ", assistantItem.Content)
	}
}
