package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	conf, err := loadConfig("config.local.json")
	if err != nil {
		fmt.Println("fail to load config: ", err)
		return
	}

	messages := []Message{
		{
			Role:    "system",
			Content: "You are Mary. You are a lovely girl.",
		},
	}

	tools := initTools()
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

		for {
			respMsg, err := triggeredByMessages(client, conf, messages, *tools)
			if err != nil {
				fmt.Println(err)
				break
			}

			messages = append(messages, *respMsg)

			for _, toolCall := range respMsg.ToolCalls {
				fmt.Println("ToolCall: ", toolCall)
				resultMsg := execute(toolCall)
				fmt.Println("ToolCallResult: ", resultMsg)
				messages = append(messages, resultMsg)
			}

			if len(respMsg.ToolCalls) == 0 {
				fmt.Println("Assistant: ", respMsg.Content)
				break
			}
		}
	}
}
