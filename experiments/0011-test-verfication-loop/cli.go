package main

import (
	"fmt"
)

func runCli(task string, configDir string) {
	conf, err := loadConfig(configDir)
	if err != nil {
		fmt.Println("fail to load config: ", err)
		return
	}

	provider, err := NewProvider(&conf.LLMProviderConfig)
	if err != nil {
		fmt.Println("fail to init provider: ", err)
		return
	}

	runner := NewRunner(conf.LLMProviderConfig.Model, provider, conf.AgentRunConfig)
	assistantItem, err := runner.Run(UserItem{Content: task})

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Assistant: ", assistantItem.Content)
	}
}
