package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

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

func Request() error {
	m := Message{
		Role:    "user",
		Content: "capital of the France",
	}
	cr := ChatRequest{
		Model:    os.Getenv("MODEL_NAME"),
		Messages: []Message{m},
		Stream:   false,
		ChatTemplateKwargs: ChatTemplateKwargs{
			EnableThinking: false,
		},
	}

	b, err := json.Marshal(cr)
	if err != nil {
		return fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.TODO(), "POST", os.Getenv("LLM_URL"), bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("new request to llm: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("LLM_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{} //nolint:exhaustruct // it's ok for pet project
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("make request to llm: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("llm request status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	chatResponse := ChatResponse{
		Choices: []Choice{},
	}

	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
		return fmt.Errorf("unmarshal response body to chat request: %w", err)
	}

	fmt.Println("llm response: ", chatResponse.Choices[0].Message.Content)

	return nil
}
