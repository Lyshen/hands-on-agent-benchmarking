package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

type LLMClient struct {
	Conf *Config
    Client *http.Client
	Adapter *Adapter
}

func (client *LLMClient) Init(conf *Config) {
	client.Conf = conf
	client.Client = &http.Client{}
	client.Adapter = &Adapter{}
}

func (client LLMClient) send(reqBytes []byte) ([]byte, error) {
    body := bytes.NewReader(reqBytes)
    req, err := http.NewRequest("POST", client.Conf.URL, body)
    if err != nil {
        return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer " + client.Conf.APIKey)

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

func (client LLMClient) triggeredBy(contexter Contexter) (*Message, error) {
	req, err := client.Adapter.convert(contexter)
	if err != nil {
		return nil, err
	}

	respBytes, err := client.send(req)
	if err != nil {
		return nil, err
	}

	respMsg, err := client.Adapter.unconvert(respBytes)
	if err != nil {
		return nil, err
	}

	return respMsg, nil
}
