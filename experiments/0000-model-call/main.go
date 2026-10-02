package main

import (
	"bytes"
	"encoding/json"
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

type OpenAIRequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Message Message `json:"message"`
}

type OpenAIResponseBody struct {
	Choices []Choice `json:"choices"`
}

func main() {
	data, err := os.ReadFile("config.local.json")

	if err != nil {
		fmt.Println(err)
		return
	}

	var conf Config
	err = json.Unmarshal(data, &conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(conf.URL)
	fmt.Println(conf.Model)

	message := Message{
		Role:    "user",
		Content: "Hello",
	}
	messages := []Message{message}
	requestBody := OpenAIRequestBody{
		Model:    conf.Model,
		Messages: messages,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println(err)
		return
	}

	reader := bytes.NewReader(body)

	req, err := http.NewRequest(
		"POST",
		conf.URL,
		reader,
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+conf.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	respStr := string(respBytes)
	fmt.Println(resp.Status)
	fmt.Println(respStr)

	var respChoice OpenAIResponseBody
	err = json.Unmarshal(respBytes, &respChoice)
	if err != nil {
		fmt.Println("response parse error")
		return 
	}
	fmt.Println(respChoice.Choices[0].Message.Content)
}
