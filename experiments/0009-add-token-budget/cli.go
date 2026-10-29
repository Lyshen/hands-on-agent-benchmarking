package main

import (
    "fmt"
	"os"
)

func loadConfigFromEnv() (LLMProviderConfig, error) {
    url := os.Getenv("OPENAI_URL")
    api_key := os.Getenv("OPENAI_API_KEY")
    model := os.Getenv("OPENAI_MODEL")

	if url == "" {
        return LLMProviderConfig{}, fmt.Errorf("OPENAI_URL is required")
	}
	if api_key == "" {
        return LLMProviderConfig{}, fmt.Errorf("OPENAI_API_KEY is required")
	}
	if model == "" {
        return LLMProviderConfig{}, fmt.Errorf("OPENAI_MODELL is required")
	}

	return LLMProviderConfig {
        URL : url,
		APIKey : api_key,
		Model : model,
		Protocol : "openai",
	}, nil
}

func runCli(task string) {
	conf, err := loadConfigFromEnv()
	if err != nil {
		fmt.Println("fail to load config: ", err)
		return
	}

	provider, err := NewProvider(&conf)
	if err != nil {
		fmt.Println("fail to init provider: ", err)
		return
	}
	state := Context{}
	state.Init(conf.Model)

	userItem := UserItem{Content: task}

	state.Add(userItem)

    assistantItem, err := runTurn(&state, provider, nil)
	
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Assistant: ", assistantItem.Content)
	}
}
