package main

func toHarnessInput(message LarkMessage) UserItem {
	return UserItem{
		Content: message.TextMessage.Text,
	}
}

func toLarkMessage(chatID string, item AssistantItem) LarkMessage {
	return LarkMessage{
		ChatId: chatID,
		TextMessage: TextMessage{
			Text: item.Content,
		},
	}
}
