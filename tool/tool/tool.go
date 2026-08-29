package tool

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
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

type WeatherArgs struct {
	City string `json:"city"`
}

func WeatherTool() Tool {
	return Tool{
		Type: "function",
		Function: Function{
			Name:        "get_weather",
			Description: "Get the current weather for a city.",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"city": {
						Type:        "string",
						Description: "The city, e.g. San Francisco.",
					},
				},
				Required: []string{"city"},
			},
		},
	}
}
