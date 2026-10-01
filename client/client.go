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
)

const httpTimeout = 30 * time.Second

var (
	ErrToolCallsCorrupted = errors.New("it seems tool calls tokens arrived corrupted")
	errStatusCode         = errors.New("llm request status code")
	errStatusBadRequest   = errors.New("status 400 from server")
)

type Client struct {
	http  http.Client
	cfg   *config.Config
	Tools []tool.Tool
}

func New(cfg *config.Config, tools ...tool.Tool) *Client {
	c := http.Client{ //nolint:exhaustruct // it's ok for petproject
		Timeout: httpTimeout,
	}

	return &Client{
		http:  c,
		cfg:   cfg,
		Tools: tools,
	}
}

func (c *Client) Request(ctx context.Context, history []Message) (ChatResponse, error) {
	cr := ChatRequest{
		Model:    c.cfg.ModelName,
		Messages: history,
		Stream:   false,
		ChatTemplateKwargs: ChatTemplateKwargs{
			EnableThinking: false,
		},
		Tools:      c.Tools,
		ToolChoice: "auto",
	}

	b, err := json.Marshal(cr)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.LLMURL, bytes.NewBuffer(b))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("new request to llm: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("make request to llm: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusBadRequest {
			return ChatResponse{}, fmt.Errorf("%w, body: %s", errStatusBadRequest, body)
		}

		return ChatResponse{}, fmt.Errorf("%w: %d, body: %s", errStatusCode, res.StatusCode, body)
	}

	chatResponse := ChatResponse{
		Choices: []Choice{},
	}

	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("unmarshal response body to chat request: %w", err)
	}

	return chatResponse, nil
}
