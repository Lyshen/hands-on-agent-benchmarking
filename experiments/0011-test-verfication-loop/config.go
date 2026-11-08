package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	LarkConnectorConfig LarkConnectorConfig `json:"lark_connector_config"`
	LLMProviderConfig   LLMProviderConfig   `json:"llm_provider_config"`
	AgentRunConfig      AgentRunConfig      `json:"agent_run_config"`
}

type LoopConfig struct {
	SystemPrompt        string `json:"system_prompt"`
	TokenBudget         *int   `json:"token_budget"`
	MaxStepsPerTurn     *int   `json:"max_steps_per_turn"`
	ShellTimeoutSeconds int    `json:"shell_timeout_seconds"`
}

type AgentRunConfig struct {
	MainLoop         LoopConfig `json:"main_loop"`
	VerificationLoop LoopConfig `json:"verification_loop"`

	MaxReviews int `json:"max_reviews"`
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

	if conf.AgentRunConfig.MaxReviews < 0 {
		return conf, fmt.Errorf("maxreviews of agent run conf is negative :%d", conf.AgentRunConfig.MaxReviews)
	}

	if conf.AgentRunConfig.MainLoop.ShellTimeoutSeconds <= 0 {
		return conf, fmt.Errorf("ShellTimeoutSeconds of MainLoop needs to be positive :%d", conf.AgentRunConfig.MainLoop.ShellTimeoutSeconds)
	}

	if conf.AgentRunConfig.VerificationLoop.ShellTimeoutSeconds <= 0 {
		return conf, fmt.Errorf("ShellTimeoutSeconds of VerificationLoop needs to be positive :%d", conf.AgentRunConfig.VerificationLoop.ShellTimeoutSeconds)
	}

	return conf, nil
}
