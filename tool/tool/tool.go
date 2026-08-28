package tool

type City struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Properties struct {
	City City `json:"city"`
}

type Parameters struct {
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
	Required   []string   `json:"required"`
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

func New() Tool {
	return Tool{
		Type: "function",
		Function: Function{
			Name:        "get_weather",
			Description: "Get the current weather for a city.",
			Parameters: Parameters{
				Type: "object",
				Properties: Properties{
					City: City{
						Type:        "string",
						Description: "The city, e.g. San Francisco.",
					},
				},
				Required: []string{"city"},
			},
		},
	}
}

// "tools": [
//   {
//     "type": "function",
//     "function": {
//       "name": "get_weather",
//       "description": "Get the current weather for a city.",
//       "parameters": {
//         "type": "object",
//         "properties": {
//           "location": {
//             "type": "string",
//             "description": "The city and state, e.g. San Francisco, CA"
//           }
//         },
//         "required": ["location"]
//       }
//     }
//   }
// ],
// "tool_choice": "auto"
