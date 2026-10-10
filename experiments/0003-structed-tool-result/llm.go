package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func buildRequest(conf Config, messages []Message, tools []Tool) (*http.Request, error) {
	chatReq := ChatRequest{
		Model:    conf.Model,
		Messages: messages,
		Tools:    tools,
	}

	var reqBytes []byte
	reqBytes, err := json.Marshal(chatReq)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(reqBytes)

	req, err := http.NewRequest("POST", conf.URL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+conf.APIKey)

	return req, nil
}

func sendRequest(client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
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

func triggeredByMessages(client *http.Client, conf Config, messages []Message, tools []Tool) (*Message, error) {
	req, err := buildRequest(conf, messages, tools)
	if err != nil {
		return nil, err
	}

	respBytes, err := sendRequest(client, req)
	if err != nil {
		return nil, err
	}

	respMsg, err := parseResponse(respBytes)
	if err != nil {
		return nil, err
	}

	return respMsg, nil
}
