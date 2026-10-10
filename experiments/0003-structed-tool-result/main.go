package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	conf, err := loadConfig("config.local.json")
	if err != nil {
		fmt.Println("fail to load config: ", err)
		return
	}

	contexter := Contexter{}
	contexter.Init()

	client := LLMClient{}
	client.Init(&conf)

	reader := bufio.NewReader(os.Stdin)
	for {
		contexter.Output()
		fmt.Println("____________________")
		fmt.Print("You: ")

		userMsg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("fail to read user message ", err)
			continue
		}
		userMsg = strings.Trim(userMsg, "\n")

		contexter.Add(Message{Role: "user", Content: userMsg})

		for {
			respMsg, err := client.triggeredByMessages(contexter)
			if err != nil {
				fmt.Println(err)
				break
			}

			contexter.Add(*respMsg)

			for _, toolCall := range respMsg.ToolCalls {
				fmt.Println("ToolCall: ", toolCall)
				resultMsg := execute(toolCall)
				fmt.Println("ToolCallResult: ", resultMsg)
				contexter.Add(resultMsg)
			}

			if len(respMsg.ToolCalls) == 0 {
				fmt.Println("Assistant: ", respMsg.Content)
				break
			}
		}
	}
}
