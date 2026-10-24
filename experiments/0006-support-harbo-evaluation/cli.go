package main

import (
    "fmt"
)

func runCli(task string) {
	conf, err := loadConfig("conf/config.local.json")
	if err != nil {
		fmt.Println("fail to load config: ", err)
		return
	}

	provider, err := NewProvider(&conf.LLMProviderConfig)
	if err != nil {
		fmt.Println("fail to init provider: ", err)
		return
	}
	state := Context{}
	state.Init(conf.LLMProviderConfig.Model)

	userItem := UserItem{Content: task}

	state.Add(userItem)

    assistantItem, err := runTurn(&state, provider)
	
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Assistant: ", assistantItem.Content)
	}
}
