package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/paveldroo/go-agent/client"
	"github.com/paveldroo/go-agent/config"
	"github.com/paveldroo/go-agent/tool/tool"
)

var (
	errMissedArgument = errors.New(`missing argument, usage: <go-agent "what is weather in Paris?">`)
	errDefineTask     = errors.New("you should define a task for agent")
	errNoToolFound    = errors.New("no tool was found by name")
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "go-agent: error: %s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	if len(os.Args) < 2 { //nolint:mnd // convenient args usage
		return errMissedArgument
	}

	prompt := os.Args[1]

	if len(prompt) == 0 {
		return errDefineTask
	}

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	c := client.New(cfg, tool.WeatherTool())

	ctx := context.Background()

	message := client.Message{
		Role:       "user",
		Content:    prompt,
		ToolCalls:  nil,
		ToolCallID: "",
	}

	for {
		resp, err := c.Request(ctx, message)
		if err != nil {
			return fmt.Errorf("requesting llm: %w", err)
		}

		if resp.FinishReason != client.ReasonToolCalls {
			fmt.Fprintln(os.Stdout, resp.Content)

			break
		}

		message, err = handleToolCall(c, resp)
		if err != nil {
			return fmt.Errorf("process response: %w", err)
		}
	}

	return nil
}

func handleToolCall(c *client.Client, resp *client.LLMResponse) (client.Message, error) {
	if len(resp.ToolCalls) == 0 {
		return client.Message{}, fmt.Errorf("%w: reason tool calls, but no tool calls objects in response", client.ErrToolCallsCorrupted)
	}

	toolCall := resp.ToolCalls[0]
	toolName := toolCall.Function.Name

	for _, clientTool := range c.Tools {
		if clientTool.Function.Name == toolName {
			var args tool.WeatherArgs
			err := toolCall.Args(&args)
			if err != nil {
				return client.Message{}, fmt.Errorf("parse tool call args: %w", err)
			}

			callRes := clientTool.Exec(args.City)

			return client.Message{
				Role:       "tool",
				Content:    callRes,
				ToolCalls:  nil,
				ToolCallID: toolCall.ID,
			}, nil
		}
	}

	return client.Message{}, fmt.Errorf("%w: %s", errNoToolFound, toolName)
}
