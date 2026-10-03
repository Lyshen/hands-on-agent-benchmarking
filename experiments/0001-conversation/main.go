package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	URL    string `json:"url"`
	APIKey string `json:"apikey"`
	Model  string `json:"model"`
}

func loadConfig(fileName string) (Config, error) {
	data, err := os.ReadFile(fileName)
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

func buildRequest(conf Config, messages []Message) (*http.Request, error) {
	chatReq := ChatRequest{
		Model:    conf.Model,
		Messages: messages,
	}
	chatReqBytes, err := json.Marshal(chatReq)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(chatReqBytes)
	req, err := http.NewRequest("POST", conf.URL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+conf.APIKey)
	return req, nil
}

func sendRequest(req http.Request) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(&req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("fail to get model response " + resp.Status)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBytes, nil
}

func parseResponse(respBytes []byte) (Message, error) {
	var chatResp ChatResponse
	err := json.Unmarshal(respBytes, &chatResp)
	if err != nil {
		return Message{}, err
	}

	if len(chatResp.Choices) == 0 {
		return Message{}, errors.New("no found in response choices")
	}

	return chatResp.Choices[0].Message, nil
}

func main() {
	conf, err := loadConfig("config.local.json")
	if err != nil {
		fmt.Println("fail to load config ", err)
		return
	}

	messages := []Message{
		Message{
			Role:    "system",
			Content: "You are Bobby. Chat with user.",
		},
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("You:")

		var userMsg string
		userMsg, err := reader.ReadString(byte('\n'))
		if err != nil {
			fmt.Println("fail to get user message: ", err)
			continue
		}
		userMsg = strings.Trim(userMsg, "\n")

		if userMsg == "exit" {
			break
		}

		messages = append(messages, Message{Role: "user", Content: userMsg})

		req, err := buildRequest(conf, messages)
		if err != nil {
			fmt.Println("fail to build request: ", err)
			continue
		}

		respBytes, err := sendRequest(*req)
		if err != nil {
			fmt.Println("fail to send request: ", err)
			continue
		}

		assistantMsg, err := parseResponse(respBytes)
		if err != nil {
			fmt.Println("fail to parse response: ", err)
			continue
		}

		fmt.Println("Assistant:", assistantMsg.Content)
		messages = append(messages, assistantMsg)
	}
}
