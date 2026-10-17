package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	URL    string `json:"url"`
	APIKey string `json:"apikey"`
	Model  string `json:"model"`
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
