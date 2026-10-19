package main

import (
	"errors"
)

type Provider interface {
	Generate(state *Context) (TranscriptItem, error)
}

func NewProvider(conf *Config) (Provider, error) {
	switch conf.Protocol {
	case "openai":
		provider := &OpenAIProvider{}
		provider.Init(conf)
		return provider, nil
	case "ollama":
		provider := &OllamaProvider{}
		provider.Init(conf)
		return provider, nil
	default:
		return nil, errors.New("Provider protocol is not supported")
	}
}
