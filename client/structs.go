package client

import (
	"github.com/paveldroo/go-agent/tool/tool"
	"github.com/paveldroo/go-agent/tool/tool_call"
)

type Message struct {
	Role      string               `json:"role"`
	Content   string               `json:"content"`
	ToolCalls []tool_call.ToolCall `json:"tool_calls"`
}

type ChatTemplateKwargs struct {
	EnableThinking bool `json:"enable_thinking"`
}

type ChatRequest struct {
	Model              string             `json:"model"`
	Messages           []Message          `json:"messages"`
	Stream             bool               `json:"stream"`
	ChatTemplateKwargs ChatTemplateKwargs `json:"chat_template_kwargs"`
	Tools              []tool.Tool        `json:"tools"`
	ToolChoice         string             `json:"tool_choice"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}
