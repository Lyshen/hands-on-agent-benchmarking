package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Config struct {
	URL    string `json:"url"`
	APIKey string `json:"apikey"`
	Model  string `json:"model"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

func loadConfig(fileName string) (Config, error) {
	conf := Config{}
	data, err := os.ReadFile(fileName)
	if err != nil {
		return conf, err
	}

	err = json.Unmarshal(data, &conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

func buildRequest(conf Config) (*http.Request, error) {
	message := Message{
		Role:    "user",
		Content: "Hello",
	}
	chatReq := ChatRequest{
		Model:    conf.Model,
		Messages: []Message{message},
	}

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

func sendRequest(req *http.Request) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("model request failed:" + resp.Status)
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

func main() {
	conf, err := loadConfig("config.local.json")
	if err != nil {
		fmt.Println("Fail to load conf file: ", err)
		return
	}

	req, err := buildRequest(conf)
	if err != nil {
		fmt.Println("Fail to build hello request: ", err)
		return
	}

	resp, err := sendRequest(req)
	if err != nil {
		fmt.Println("Fail to get response: ", err)
		return
	}

	fmt.Println(string(resp))

	message, err := parseResponse(resp)
	if err != nil {
		fmt.Println("Fail to parse response: ", err)
		return
	}

	fmt.Println(message.Content)
}
