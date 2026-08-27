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
)

var (
	errStatusCode = errors.New("llm request status code")
	errNoContent  = errors.New("no content from llm")
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
		Role:    "user",
		Content: prompt,
	}
	cr := ChatRequest{
		Model:    c.cfg.ModelName,
		Messages: []Message{m},
		Stream:   false,
		ChatTemplateKwargs: ChatTemplateKwargs{
			EnableThinking: false,
		},
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

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %d", errStatusCode, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
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

	return chatResponse.Choices[0].Message.Content, nil
}
