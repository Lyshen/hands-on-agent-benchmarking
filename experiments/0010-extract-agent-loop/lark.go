package main

import (
	"context"
	"fmt"
)

func sendOutMessage(connector LarkConnector, chatId string, item AssistantItem) {
	connector.OutMessages <- LarkMessage{
		ChatId: chatId,
		TextMessage: TextMessage{
			Text: item.Content,
		},
	}
}

func readInMessage(message LarkMessage) UserItem {
	return UserItem{
		Content: message.TextMessage.Text,
	}
}

func runLark() {
	conf, err := loadConfig("conf/config.local.json")
	if err != nil {
		fmt.Println("fail to load config: ", err)
		return
	}

	provider, err := NewProvider(&conf.LLMProviderConfig)
	if err != nil {
		fmt.Println("fail to init provider: ", err)
		return
	}

	state := NewContext(conf.LLMProviderConfig.Model)
	runner := NewRunner(NewLoop(&state, provider, conf.AgentRunConfig.TokenBudget))

	connectorErr := make(chan error, 1)
	connector := NewLarkConnector(conf.LarkConnectorConfig)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		fmt.Println("starting lark connector...")
		connectorErr <- connector.start(ctx)
	}()

	for {
		select {
		case err := <-connectorErr:
			fmt.Println("start connector failed:", err)
			return

		case inMessage := <-connector.InMessages:
			fmt.Print("\n\nYou: ", inMessage.TextMessage.Text)
			fmt.Println()

			userItem := readInMessage(inMessage)
			state.Add(userItem)
			assistantItem, err := runner.Run()

			if err != nil {
				fmt.Println(err)
				sendOutMessage(connector, inMessage.ChatId, AssistantItem{Content: "Inner error, please try it again."})
			} else {
				fmt.Println("Assistant: ", assistantItem.Content)
				sendOutMessage(connector, inMessage.ChatId, assistantItem)
			}

		case <-ctx.Done():
			fmt.Println("context is cancel")
			return
		}
	}
}
