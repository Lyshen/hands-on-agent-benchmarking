package main

import (
	"context"
	"fmt"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

type LarkConnector struct {
	wsClient *larkws.Client
}

func handleLarkMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	fmt.Println("received lark message event")
	if event == nil ||
		event.Event == nil ||
		event.Event.Message == nil {
		return nil
	}

	message := event.Event.Message
	if message.ChatId == nil || message.Content == nil {
		return nil
	}

	fmt.Println("chatID:", *message.ChatId)
	fmt.Println("content:", *message.Content)

	return nil
}

func NewLarkConnector(conf LarkConnectorConfig) LarkConnector {
	handler := larkevent.NewEventDispatcher("", "").OnP2MessageReceiveV1(handleLarkMessage)
	connector := LarkConnector{}
	connector.wsClient = larkws.NewClient(conf.AppID, conf.AppSecret, larkws.WithEventHandler(handler))

	return connector
}

func (connector *LarkConnector) start(ctx context.Context) error {
	return connector.wsClient.Start(ctx)
}
