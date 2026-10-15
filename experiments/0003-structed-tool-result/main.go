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
		userItem := Item{
			Kind: "user",
			UserItem: &UserItem{
				Content: userMsg,
			},
		}

		contexter.Add(userItem)

		var assistantItem Item
		for {
			req := adapter.buildOpenAIRequest(contexter)
			resp, err := provider.send(req)
			if err != nil {
				fmt.Println(err)
				break
			}
			respMsg := adapter.buildMessage(resp)

			if len(respMsg.ToolCalls) == 0 {
				assistantItem = Item{
					Kind: "assistant",
					AssistantItem: &AssistantItem{
						Content: respMsg.Content,
					},
				}
				contexter.Add(assistantItem)
				break
			}

			for _, toolCall := range respMsg.ToolCalls {
				toolInvocation := adapter.buildToolInvocation(toolCall)
				toolInvocationItem := Item{
					Kind:           "toolInvocation",
					ToolInvocation: &toolInvocation,
				}
				contexter.Add(toolInvocationItem)

				toolResult := execute(toolInvocation)
				toolResultItem := Item{
					Kind:       "toolResult",
					ToolResult: &toolResult,
				}
				contexter.Add(toolResultItem)
			}
		}

		fmt.Println("Assistant: ", assistantItem.AssistantItem.Content)
	}
}
