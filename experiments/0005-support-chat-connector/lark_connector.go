package main

import (
	"context"
	"encoding/json"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

type InMessage struct {
	ChatId  string
	Content string
}

type LarkInMessage struct {
	Text string `json:"text"`
}

type LarkConnector struct {
	InMessages chan InMessage
	wsClient   *larkws.Client
}

func (connector *LarkConnector) handleLarkMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	if event == nil ||
		event.Event == nil ||
		event.Event.Message == nil {
		return nil
	}

	message := event.Event.Message
	if message.ChatId == nil || message.Content == nil {
		return nil
	}

	var data LarkInMessage
	err := json.Unmarshal([]byte(*message.Content), &data)
	if err != nil {
		return err
	}

	inMessage := InMessage{
		ChatId:  *message.ChatId,
		Content: data.Text,
	}
	connector.InMessages <- inMessage

	return nil
}

func NewLarkConnector(conf LarkConnectorConfig) LarkConnector {
	connector := LarkConnector{
		InMessages: make(chan InMessage, 10),
	}
	handler := larkevent.NewEventDispatcher("", "").OnP2MessageReceiveV1(connector.handleLarkMessage)
	connector.wsClient = larkws.NewClient(conf.AppID, conf.AppSecret, larkws.WithEventHandler(handler))

	return connector
}

func (connector *LarkConnector) start(ctx context.Context) error {
	return connector.wsClient.Start(ctx)
}
