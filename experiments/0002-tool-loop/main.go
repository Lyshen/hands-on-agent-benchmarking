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

func loadConfig(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
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

type ToolCallSchema struct {
	Arguments string `json:"arguments"`
	Name      string `json:"name"`
}

type ToolCall struct {
	Function ToolCallSchema `json:"function"`
	Id       string         `json:"id"`
	Type     string         `json:"type"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type Command struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Properties struct {
	Command Command `json:"command"`
}

type Parameters struct {
	Type                 string     `json:"type"`
	Properties           Properties `json:"properties"`
	Required             []string   `json:"required"`
	AdditionalProperties bool       `json:"additionalProperties"`
}

type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

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
		messages = messages[:len(messages)-1]
		return nil, err
	}

	respBytes, err := sendRequest(client, req)
	if err != nil {
		messages = messages[:len(messages)-1]
		return nil, err
	}

	fmt.Println(string(respBytes))

	assistantMsg, err := parseResponse(respBytes)
	if err != nil {
		messages = messages[:len(messages)-1]
		return nil, err
	}

	return assistantMsg, nil

}

func execute(toolCall ToolCall) Message {
	message := Message{
		Role:       "tool",
		ToolCallID: toolCall.Id,
		Content:    "OK",
	}

	return message
}

func main() {
	conf, err := loadConfig("config.local.json")
	if err != nil {
		fmt.Println("fail to load config: ", err)
		return
	}
	fmt.Println(conf.URL)
	messages := []Message{
		{
			Role:    "system",
			Content: "You are Mary. You are a lovely girl.",
		},
	}
	fmt.Println(messages[0].Content)
	tools := []Tool{
		{
			Type: "function",
			Function: Function{
				Name:        "shell",
				Description: "Run an allowed read-only command in the journal workspace.",
				Parameters: Parameters{
					Type: "object",
					Properties: Properties{
						Command: Command{
							Type:        "string",
							Description: "The command to run.",
						},
					},
					Required: []string{
						"command",
					},
					AdditionalProperties: false,
				},
			},
		},
	}
	fmt.Println(tools[0].Type)
	reader := bufio.NewReader(os.Stdin)
	client := &http.Client{}
	for {
		fmt.Println(messages)
		fmt.Println("____________________")
		fmt.Print("You: ")

		userMsg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("fail to read user message ", err)
			continue
		}
		userMsg = strings.Trim(userMsg, "\n")

		messages = append(messages, Message{Role: "user", Content: userMsg})

		assistantMsg, err := triggeredByMessages(client, conf, messages, tools)
		if err != nil {
			fmt.Println(err)
			continue
		}

		messages = append(messages, *assistantMsg)
		if len(assistantMsg.ToolCalls) != 0 {
			fmt.Println("Assistant (ToolCall): ", assistantMsg.ToolCalls[0])
			resultMsg := execute(assistantMsg.ToolCalls[0])
			messages = append(messages, resultMsg)

			assistantMsg, err = triggeredByMessages(client, conf, messages, tools)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Assistant: ", assistantMsg.Content)
			messages = append(messages, *assistantMsg)

		} else {
			fmt.Println("Assistant: ", assistantMsg.Content)
		}
	}
}
