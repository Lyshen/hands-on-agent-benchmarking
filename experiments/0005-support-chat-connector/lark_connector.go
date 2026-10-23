package main

import (
	"context"
	"encoding/json"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

type LarkMessage struct {
	ChatId      string
	TextMessage TextMessage
}

type TextMessage struct {
	Text string `json:"text"`
}

type LarkConnector struct {
	InMessages  chan LarkMessage
	OutMessages chan LarkMessage
	wsClient    *larkws.Client
	apiClient   *lark.Client
}

func NewLarkConnector(conf LarkConnectorConfig) LarkConnector {
	connector := LarkConnector{
		InMessages:  make(chan LarkMessage, 10),
		OutMessages: make(chan LarkMessage, 10),
	}

	handler := larkevent.NewEventDispatcher("", "").OnP2MessageReceiveV1(connector.handleLarkMessage)
	connector.wsClient = larkws.NewClient(conf.AppID, conf.AppSecret, larkws.WithEventHandler(handler))
	connector.apiClient = lark.NewClient(conf.AppID, conf.AppSecret)

	return connector
}

func (connector *LarkConnector) start(ctx context.Context) error {
	go connector.runSender(ctx)
	return connector.wsClient.Start(ctx)
}

func (connector *LarkConnector) sendMessage(ctx context.Context, message LarkMessage) error {
	contentBytes, err := json.Marshal(message.TextMessage)
	if err != nil {
		return err
	}

	body := larkim.NewCreateMessageReqBodyBuilder().ReceiveId(message.ChatId).MsgType("text").Content(string(contentBytes)).Build()
	req := larkim.NewCreateMessageReqBuilder().ReceiveIdType("chat_id").Body(body).Build()

	resp, err := connector.apiClient.Im.V1.Message.Create(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Success() {
		return fmt.Errorf("send message failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

func (connector *LarkConnector) runSender(ctx context.Context) {
	for {
		select {
		case message := <-connector.OutMessages:
			err := connector.sendMessage(ctx, message)
			if err != nil {
				continue
			}
		case <-ctx.Done():
			return
		}
	}
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

	var data TextMessage
	err := json.Unmarshal([]byte(*message.Content), &data)
	if err != nil {
		return err
	}

	inMessage := LarkMessage{
		ChatId:      *message.ChatId,
		TextMessage: data,
	}
	connector.InMessages <- inMessage

	return nil
}
