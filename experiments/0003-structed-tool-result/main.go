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
	contexter.Init(conf.Model)

	adapter := OpenAIAdapter{}

	provider := LLMProvider{}
	provider.Init(&conf)

	reader := bufio.NewReader(os.Stdin)
	for {
		//contexter.Output()
		fmt.Print("\n\nYou: ")

		userMsg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("fail to read user message ", err)
			continue
		}
		userMsg = strings.Trim(userMsg, "\n")

		contexter.Add(Message{Role: "user", Content: userMsg})

		for {
			req := adapter.buildOpenAIRequest(contexter)
			resp, err := provider.send(req)
			if err != nil {
				fmt.Println(err)
				break
			}

			respMsg := adapter.buildMessage(resp)
			contexter.Add(respMsg)

			for _, toolCall := range respMsg.ToolCalls {
				toolInvocation := adapter.buildToolInvocation(toolCall)
				toolResult := execute(toolInvocation)
				//fmt.Println("\nToolCall: ", toolCall)
				//fmt.Println("ToolCallResult: ", resultMsg)

				contexter.Add(adapter.buildOpenAIToolResultMessage(toolResult))
			}

			if len(respMsg.ToolCalls) == 0 {
				fmt.Println("Assistant: ", respMsg.Content)
				break
			}
		}
	}
}
