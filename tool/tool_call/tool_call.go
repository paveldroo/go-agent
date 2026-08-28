package tool_call

type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCall struct {
	Function Function `json:"function"`
}
