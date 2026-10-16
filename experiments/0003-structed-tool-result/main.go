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

	context := Context{}
	context.Init(conf.Model)

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

		context.Add(UserItem{Content: userMsg})

		var assistantItem AssistantItem
	modelLoop:
		for {
			req := adapter.buildOpenAIRequest(context)
			resp, err := provider.send(req)
			if err != nil {
				fmt.Println(err)
				break modelLoop
			}

			item, err := adapter.buildTranscriptItem(resp)
			if err != nil {
				fmt.Println(err)
				break modelLoop
			}

			switch value := item.(type) {
			case AssistantItem:
				context.Add(value)
				assistantItem = value
				break modelLoop

			case ToolCallsItem:
				context.Add(value)
				for _, toolInvocation := range value.Calls {
					toolResult := execute(toolInvocation)
					context.Add(ToolResultItem{Result: toolResult})
				}
			}
		}

		fmt.Println("Assistant: ", assistantItem.Content)
	}
}
