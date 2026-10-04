package conversation

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/paveldroo/go-agent/client"
	"github.com/paveldroo/go-agent/tool/tool"
	"github.com/paveldroo/go-agent/tool/tool_call"
)

var (
	ErrTruncated        = errors.New("response from llm is truncated")
	errUnexpectedReason = errors.New("unexpected finish reason")
	errNoChoices        = errors.New("no choices from llm")
	errNoToolFound      = errors.New("no tool was found by name")
)

type LLMClient interface {
	Request(ctx context.Context, history []client.Message) (client.ChatResponse, error)
	Tools() []tool.Tool
}

type LLMResponse struct {
	FinishReason string
	Content      string
	ToolCalls    []tool_call.ToolCall
}

type Conversation struct {
	History []client.Message
}

func New() *Conversation {
	return &Conversation{
		History: []client.Message{},
	}
}

func (c *Conversation) Run(ctx context.Context, llmClient LLMClient, prompt string) error {
	message := client.Message{
		Role:             "user",
		Content:          prompt,
		ToolCalls:        nil,
		ToolCallID:       "",
		ReasoningContent: nil,
	}

	for {
		c.History = append(c.History, message)
		resp, err := llmClient.Request(ctx, c.History)
		if err != nil {
			return fmt.Errorf("requesting llm: %w", err)
		}

		llmResp, err := c.processRes(resp)
		if err != nil {
			return fmt.Errorf("process llm client response: %w", err)
		}

		if llmResp.FinishReason != client.ReasonToolCalls {
			fmt.Fprintln(os.Stdout, llmResp.Content)

			break
		}

		message, err = handleToolCall(llmClient, llmResp)
		if err != nil {
			return fmt.Errorf("process response: %w", err)
		}
	}

	return nil
}

func (c *Conversation) processRes(cr client.ChatResponse) (*LLMResponse, error) {
	if len(cr.Choices) == 0 {
		return nil, errNoChoices
	}

	firstChoice := cr.Choices[0]

	if firstChoice.FinishReason == client.ReasonLength {
		if len(firstChoice.Message.ToolCalls) != 0 {
			return nil, fmt.Errorf("%w: %v", ErrTruncated, firstChoice.Message.ToolCalls[0].Function.Arguments)
		}

		return nil, fmt.Errorf("%w: %v", ErrTruncated, firstChoice.Message.Content)
	}

	if firstChoice.FinishReason == client.ReasonStop || firstChoice.FinishReason == client.ReasonToolCalls {
		c.History = append(c.History, firstChoice.Message)

		return &LLMResponse{
			FinishReason: firstChoice.FinishReason,
			Content:      firstChoice.Message.Content,
			ToolCalls:    firstChoice.Message.ToolCalls,
		}, nil
	}

	return nil, fmt.Errorf("%w: %s", errUnexpectedReason, firstChoice.FinishReason)
}

func handleToolCall(c LLMClient, resp *LLMResponse) (client.Message, error) {
	if len(resp.ToolCalls) == 0 {
		return client.Message{}, fmt.Errorf("%w: reason tool calls, but no tool calls objects in response", client.ErrToolCallsCorrupted)
	}

	toolCall := resp.ToolCalls[0]
	toolName := toolCall.Function.Name

	for _, clientTool := range c.Tools() {
		if clientTool.Function.Name == toolName {
			var args tool.WeatherArgs
			err := toolCall.Args(&args)
			if err != nil {
				return client.Message{}, fmt.Errorf("parse tool call args: %w", err)
			}

			callRes := clientTool.Exec(args.City)

			return client.Message{
				Role:             "tool",
				Content:          callRes,
				ToolCalls:        nil,
				ToolCallID:       toolCall.ID,
				ReasoningContent: nil,
			}, nil
		}
	}

	return client.Message{}, fmt.Errorf("%w: %s", errNoToolFound, toolName)
}
