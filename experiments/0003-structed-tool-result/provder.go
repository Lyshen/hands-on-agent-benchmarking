package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type LLMProvider struct {
	Conf   *Config
	Client *http.Client
}

func (provider *LLMProvider) Init(conf *Config) {
	provider.Conf = conf
	provider.Client = &http.Client{}
}

func (provider LLMProvider) send(chatReq OpenAIChatRequest) (OpenAIChatResponse, error) {
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
		return chatResp, errors.New("fail to get response " + resp.Status)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return chatResp, err
	}

	err = json.Unmarshal(respBytes, &chatResp)
	if err != nil {
		return chatResp, err
	}

	if len(chatResp.Choices) == 0 {
		return chatResp, errors.New("no message found in response choices")
	}

	return chatResp, nil
}
