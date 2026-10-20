package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	larkConfig        LarkConnectorConfig `json:"lark_connector_config"`
	llmProviderConfig LLMProviderConfig   `json:"llm_provider_config"`
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
