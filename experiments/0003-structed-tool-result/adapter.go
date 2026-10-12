package main

import (
    "encoding/json"
	"errors"
)

type Adapter struct {
}

func (a Adapter) convert(contexter Contexter) ([]byte, error) {
	chatReq := ChatRequest{
		Model:    contexter.Model,
		Messages: contexter.Messages,
		Tools:    contexter.Tools,
	}

	var reqBytes []byte
	reqBytes, err := json.Marshal(chatReq)
	if err != nil {
		return nil, err
	}

	return reqBytes, nil
}

func (a Adapter) unconvert(respBytes []byte) (*Message, error) {
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
