package main

import (
	"fmt"
	"os"
	"strconv"
)

func loadConfigFromEnv() (Config, error) {
	url := os.Getenv("OPENAI_URL")
	api_key := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("OPENAI_MODEL")

	var tokenBudget *int
	tokenBudgetStr := os.Getenv("TOKEN_BUDGET")

	if url == "" {
		return Config{}, fmt.Errorf("OPENAI_URL is required")
	}
	if api_key == "" {
		return Config{}, fmt.Errorf("OPENAI_API_KEY is required")
	}
	if model == "" {
		return Config{}, fmt.Errorf("OPENAI_MODEL is required")
	}
	if tokenBudgetStr != "" {
		budget, err := strconv.Atoi(tokenBudgetStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid TOKEN_BUDGET: %w", err)
		} else if budget < 0 {
			return Config{}, fmt.Errorf("TOKEN_BUDGET must be non-negative")
		} else {
			tokenBudget = &budget
		}
	}

	config := Config{
		LLMProviderConfig: LLMProviderConfig{
			URL:      url,
			APIKey:   api_key,
			Model:    model,
			Protocol: "openai",
		},
		AgentRunConfig: AgentRunConfig{
			TokenBudget: tokenBudget,
		},
	}

	return config, nil
}

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
	state := Context{}
	state.Init(conf.LLMProviderConfig.Model)

	userItem := UserItem{Content: task}

	state.Add(userItem)

	assistantItem, err := runTurn(&state, provider, conf.AgentRunConfig.TokenBudget)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Assistant: ", assistantItem.Content)
	}
}
