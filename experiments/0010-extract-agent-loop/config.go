package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	LarkConnectorConfig LarkConnectorConfig `json:"lark_connector_config"`
	LLMProviderConfig   LLMProviderConfig   `json:"llm_provider_config"`
	AgentRunConfig      AgentRunConfig      `json:"agent_run_config"`
}

type AgentRunConfig struct {
	TokenBudget *int `json:"token_budget"`
}

type LarkConnectorConfig struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type LLMProviderConfig struct {
	URL      string `json:"url"`
	APIKey   string `json:"apikey"`
	Model    string `json:"model"`
	Protocol string `json:"protocol"`
}

func loadConfig(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var conf Config
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return Config{}, err
	}

	return conf, nil
}

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
