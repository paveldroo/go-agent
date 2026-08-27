package client

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatTemplateKwargs struct {
	EnableThinking bool `json:"enable_thinking"`
}

type ChatRequest struct {
	Model              string             `json:"model"`
	Messages           []Message          `json:"messages"`
	Stream             bool               `json:"stream"`
	ChatTemplateKwargs ChatTemplateKwargs `json:"chat_template_kwargs"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}
