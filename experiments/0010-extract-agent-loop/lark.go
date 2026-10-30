package main

import (
	"context"
	"fmt"
)

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
	runner := NewRunner(LoopConfig{
		State:       state,
		Provider:    provider,
		TokenBudget: conf.AgentRunConfig.TokenBudget,
	})

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
			fmt.Printf("You:%s\n", inMessage.TextMessage.Text)

			userItem := toHarnessInput(inMessage)
			assistantItem, err := runner.Run(userItem)

			if err != nil {
				fmt.Println(err)
				connector.OutMessages <- toLarkMessage(inMessage.ChatId, AssistantItem{Content: "Inner error, please try it again."})
			} else {
				fmt.Println("Assistant: ", assistantItem.Content)
				connector.OutMessages <- toLarkMessage(inMessage.ChatId, assistantItem)
			}

		case <-ctx.Done():
			fmt.Println("context is cancel")
			return
		}
	}
}
