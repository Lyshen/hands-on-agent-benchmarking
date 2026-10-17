package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func runTurn(state *Context, provider *LLMProvider, adapter *OpenAIAdapter) (AssistantItem, error) {
	for {
		req, err := adapter.buildOpenAIRequest(*state)
		if err != nil {
			return AssistantItem{}, err
		}

		resp, err := provider.send(req)
		if err != nil {
			return AssistantItem{}, err
		}

		item, err := adapter.buildTranscriptItem(resp)
		if err != nil {
			return AssistantItem{}, err
		}

		switch value := item.(type) {
		case AssistantItem:
			state.Add(value)
			return value, nil

		case ToolCallsItem:
			state.Add(value)
			for _, toolInvocation := range value.Calls {
				toolResult := execute(toolInvocation)
				state.Add(ToolResultItem{Result: toolResult})
			}
		}
	}
}

func readUserMessage(reader *bufio.Reader) (UserItem, error) {

	userMsg, err := reader.ReadString('\n')
	eof := errors.Is(err, io.EOF)

	if err != nil && !eof {
		return UserItem{}, err
	}

	if eof && len(userMsg) == 0 {
		return UserItem{}, errors.New("stdin is end")
	}

	userMsg = strings.TrimSuffix(userMsg, "\n")
	userMsg = strings.TrimSuffix(userMsg, "\r")

	return UserItem{Content: userMsg}, nil
}

func main() {
	conf, err := loadConfig("config.local.json")
	if err != nil {
		fmt.Println("fail to load config: ", err)
		return
	}

	state := Context{}
	state.Init(conf.Model)
	provider := LLMProvider{}
	provider.Init(&conf)
	adapter := OpenAIAdapter{}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n\nYou: ")

		userItem, err := readUserMessage(reader)
		if err != nil {
			fmt.Println("fail to read user message: ", err)
			return
		}

		state.Add(userItem)

		assistantItem, err := runTurn(&state, &provider, &adapter)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Assistant: ", assistantItem.Content)
		}
	}
}
