package main


type ShellCommand struct {
	Command string `json:"command"`
}

type ToolCallSchema struct {
	Arguments string `json:"arguments"`
	Name      string `json:"name"`
}

type ToolCall struct {
	Function ToolCallSchema `json:"function"`
	Id       string         `json:"id"`
	Type     string         `json:"type"`
}

type Command struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Properties struct {
	Command Command `json:"command"`
}

type Parameters struct {
	Type                 string     `json:"type"`
	Properties           Properties `json:"properties"`
	Required             []string   `json:"required"`
	AdditionalProperties bool       `json:"additionalProperties"`
}

type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

func initTools() *[]Tool {
	tools := []Tool{
		{
			Type: "function",
			Function: Function{
				Name:        "shell",
				Description: "Run an allowed read-only command in the journal workspace.",
				Parameters: Parameters{
					Type: "object",
					Properties: Properties{
						Command: Command{
							Type:        "string",
							Description: "The command to run.",
						},
					},
					Required: []string{
						"command",
					},
					AdditionalProperties: false,
				},
			},
		},
	}

	return &tools
}

