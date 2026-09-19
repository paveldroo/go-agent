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

const httpTimeout = 30 * time.Second

var (
	ErrTruncated          = errors.New("response from llm is truncated")
	ErrToolCallsCorrupted = errors.New("it seems tool calls tokens arrived corrupted")
	errStatusCode         = errors.New("llm request status code")
	errNoContent          = errors.New("no content from llm")
	errStatusBadRequest   = errors.New("status 400 from server")
)

type LLMResponse struct {
	FinishReason string
	Content      string
	ToolCalls    []tool_call.ToolCall
}

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

func (c *Client) Request(ctx context.Context, prompt string) (*LLMResponse, error) {
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
		return nil, fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.LLMURL, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("new request to llm: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request to llm: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusBadRequest {
			return nil, fmt.Errorf("%w, body: %s", errStatusBadRequest, body)
		}

		return nil, fmt.Errorf("%w: %d, body: %s", errStatusCode, res.StatusCode, body)
	}

	llmResponse, err := processRes(body)
	if err != nil {
		return nil, fmt.Errorf("process response: %w", err)
	}

	return llmResponse, nil
}

func processRes(body []byte) (*LLMResponse, error) {
	chatResponse := ChatResponse{
		Choices: []Choice{},
	}

	err := json.Unmarshal(body, &chatResponse)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response body to chat request: %w", err)
	}

	if len(chatResponse.Choices) == 0 {
		return nil, errNoContent
	}

	firstChoice := chatResponse.Choices[0]

	if firstChoice.FinishReason == ReasonLength {
		if len(firstChoice.Message.ToolCalls) != 0 {
			return nil, fmt.Errorf("%w: %v", ErrTruncated, firstChoice.Message.ToolCalls[0].Function.Arguments)
		}

		return nil, fmt.Errorf("%w: %v", ErrTruncated, firstChoice.Message.Content)
	}

	return &LLMResponse{
		FinishReason: firstChoice.FinishReason,
		Content:      firstChoice.Message.Content,
		ToolCalls:    firstChoice.Message.ToolCalls,
	}, nil
}
