package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/paveldroo/go-agent/config"
	"github.com/paveldroo/go-agent/tool/tool"
	"github.com/paveldroo/go-agent/tool/tool_call"
)

var (
	errStatusCode       = errors.New("llm request status code")
	errNoContent        = errors.New("no content from llm")
	errStatusBadRequest = errors.New("status 400 from server")
)

type Client struct {
	http http.Client
	cfg  *config.Config
}

func New(cfg *config.Config) *Client {
	c := http.Client{} //nolint:exhaustruct // it's ok for petproject

	return &Client{
		http: c,
		cfg:  cfg,
	}
}

func (c *Client) Request(prompt string) (string, error) {
	m := Message{
		Role:      "user",
		Content:   prompt,
		ToolCalls: []tool_call.ToolCall{},
	}

	t := tool.New()

	cr := ChatRequest{
		Model:    c.cfg.ModelName,
		Messages: []Message{m},
		Stream:   false,
		ChatTemplateKwargs: ChatTemplateKwargs{
			EnableThinking: false,
		},
		Tools:      []tool.Tool{t},
		ToolChoice: "auto",
	}

	b, err := json.Marshal(cr)
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, c.cfg.LLMURL, bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("new request to llm: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{} //nolint:exhaustruct // it's ok for pet project
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("make request to llm: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusBadRequest {
			return "", fmt.Errorf("%w, body: %s", errStatusBadRequest, body)
		}

		return "", fmt.Errorf("%w: %d", errStatusCode, res.StatusCode)
	}

	chatResponse := ChatResponse{
		Choices: []Choice{},
	}

	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
		return "", fmt.Errorf("unmarshal response body to chat request: %w", err)
	}

	if len(chatResponse.Choices) == 0 {
		return "", errNoContent
	}

	firstChoice := chatResponse.Choices[0]

	if len(firstChoice.Message.ToolCalls) != 0 {
		return printToolCallArgs(firstChoice)
	}

	return firstChoice.Message.Content, nil
}

func printToolCallArgs(choice Choice) (string, error) {
	firstToolCall := choice.Message.ToolCalls[0]

	var args struct {
		City string `json:"city"`
	}

	err := firstToolCall.Args(&args)
	if err != nil {
		return "", fmt.Errorf("unmarshal tool call args: %w", err)
	}

	return fmt.Sprintf("%s(city=%q)", firstToolCall.Function.Name, args.City), nil
}
