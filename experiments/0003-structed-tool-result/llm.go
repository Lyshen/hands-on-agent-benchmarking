package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type LLMClient struct {
	Conf *Config
    Client *http.Client
}

func (client *LLMClient) Init(conf *Config) {
	client.Conf = conf
	client.Client = &http.Client{}
}

func (client LLMClient) buildRequest(contexter Contexter) (*http.Request, error) {
	chatReq := ChatRequest{
		Model:    client.Conf.Model,
		Messages: contexter.Messages,
		Tools:    contexter.Tools,
	}

	var reqBytes []byte
	reqBytes, err := json.Marshal(chatReq)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(reqBytes)

	req, err := http.NewRequest("POST", client.Conf.URL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.Conf.APIKey)

	return req, nil
}

func (client LLMClient) send(req *http.Request) ([]byte, error) {
	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("fail to get response " + resp.Status)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBytes, nil
}

func parseResponse(respBytes []byte) (*Message, error) {
	var resp ChatResponse
	err := json.Unmarshal(respBytes, &resp)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no message found in response choices")
	}

	return &resp.Choices[0].Message, nil
}

func (client LLMClient) triggeredByMessages(contexter Contexter) (*Message, error) {
	req, err := client.buildRequest(contexter)
	if err != nil {
		return nil, err
	}

	respBytes, err := client.send(req)
	if err != nil {
		return nil, err
	}

	respMsg, err := parseResponse(respBytes)
	if err != nil {
		return nil, err
	}

	return respMsg, nil
}
