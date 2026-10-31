package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAIProvider struct {
	Conf    *LLMProviderConfig
	Client  *http.Client
	Adapter OpenAIAdapter
}

func (provider OpenAIProvider) Generate(state Context) (ProviderResponse, error) {
	req, err := provider.Adapter.buildRequest(state)
	if err != nil {
		return ProviderResponse{Item: AssistantItem{}, Usage: TokenUsage{}}, err
	}

	resp, err := provider.send(req)
	if err != nil {
		return ProviderResponse{Item: AssistantItem{}, Usage: TokenUsage{}}, err
	}

	item, err := provider.Adapter.buildTranscriptItem(resp)
	if err != nil {
		return ProviderResponse{Item: AssistantItem{}, Usage: TokenUsage{}}, err
	}

	return ProviderResponse{
		Item:  item,
		Usage: provider.Adapter.buildTokenUsage(resp),
	}, nil
}

func (provider *OpenAIProvider) Init(conf *LLMProviderConfig) {
	provider.Conf = conf
	provider.Client = &http.Client{}
	provider.Adapter = OpenAIAdapter{}
}

func (provider OpenAIProvider) send(chatReq OpenAIChatRequest) (OpenAIChatResponse, error) {
	chatResp := OpenAIChatResponse{}
	var reqBytes []byte
	reqBytes, err := json.Marshal(chatReq)
	if err != nil {
		return chatResp, err
	}

	body := bytes.NewReader(reqBytes)
	req, err := http.NewRequest("POST", provider.Conf.URL, body)
	if err != nil {
		return chatResp, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+provider.Conf.APIKey)

	resp, err := provider.Client.Do(req)
	if err != nil {
		return chatResp, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBytes, _ := io.ReadAll(resp.Body)
		return chatResp, fmt.Errorf(
			"request failed: %s: %s",
			resp.Status,
			string(respBytes),
		)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return chatResp, err
	}

	err = json.Unmarshal(respBytes, &chatResp)
	if err != nil {
		return chatResp, err
	}

	return chatResp, nil
}
