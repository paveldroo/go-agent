package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/paveldroo/go-agent/config"
	"github.com/paveldroo/go-agent/tool/tool"
	"github.com/paveldroo/go-agent/tool/tool_call"
)

const (
	reasonToolCalls = "tool_calls"
	reasonLength    = "length"
	httpTimeout     = 30 * time.Second
)

var (
	errStatusCode         = errors.New("llm request status code")
	errNoContent          = errors.New("no content from llm")
	errStatusBadRequest   = errors.New("status 400 from server")
	errToolCallsCorrupted = errors.New("it seems tool calls tokens arrived corrupted")
)

type Client struct {
	http        http.Client
	cfg         *config.Config
	weatherTool tool.Tool
}

func New(cfg *config.Config) *Client {
	c := http.Client{ //nolint:exhaustruct // it's ok for petproject
		Timeout: httpTimeout,
	}

	return &Client{
		http:        c,
		cfg:         cfg,
		weatherTool: tool.WeatherTool(),
	}
}

func (c *Client) Request(ctx context.Context, prompt string) (string, error) {
	m := Message{
		Role:      "user",
		Content:   prompt,
		ToolCalls: []tool_call.ToolCall{},
	}

	cr := ChatRequest{
		Model:    c.cfg.ModelName,
		Messages: []Message{m},
		Stream:   false,
		ChatTemplateKwargs: ChatTemplateKwargs{
			EnableThinking: false,
		},
		Tools:      []tool.Tool{c.weatherTool},
		ToolChoice: "auto",
	}

	b, err := json.Marshal(cr)
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.LLMURL, bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("new request to llm: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
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

		return "", fmt.Errorf("%w: %d, body: %s", errStatusCode, res.StatusCode, body)
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

	if firstChoice.FinishReason == reasonToolCalls ||
		(firstChoice.FinishReason == reasonLength && len(firstChoice.Message.ToolCalls) != 0) {
		return parseToolCallArgs(firstChoice)
	}

	return firstChoice.Message.Content, nil
}

func parseToolCallArgs(choice Choice) (string, error) {
	if len(choice.Message.ToolCalls) == 0 {
		return "", errToolCallsCorrupted
	}
	firstToolCall := choice.Message.ToolCalls[0]

	weatherArgs := tool.WeatherArgs{
		City: "",
	}
	err := firstToolCall.Args(&weatherArgs)
	if err != nil {
		return "", fmt.Errorf("unmarshal tool call args: %w", err)
	}

	return fmt.Sprintf("%s(city=%q)", firstToolCall.Function.Name, weatherArgs.City), nil
}
