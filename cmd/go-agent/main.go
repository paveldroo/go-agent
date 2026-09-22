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

	m := client.Message{
		Role:       "user",
		Content:    prompt,
		ToolCalls:  nil,
		ToolCallID: "",
	}

	resp, err := c.Request(ctx, m)
	if err != nil {
		return fmt.Errorf("requesting llm: %w", err)
	}

	err = processResponse(ctx, c, resp)
	if err != nil {
		return fmt.Errorf("process response: %w", err)
	}

	return nil
}

func processResponse(ctx context.Context, c *client.Client, resp *client.LLMResponse) error {
	if resp.FinishReason == client.ReasonToolCalls {
		if len(resp.ToolCalls) == 0 {
			return fmt.Errorf("%w: reason tool calls, but no tool calls objects in response", client.ErrToolCallsCorrupted)
		}

		toolCall := resp.ToolCalls[0]

		for _, clientTool := range c.Tools {
			if clientTool.Function.Name == toolCall.Function.Name {
				var args tool.WeatherArgs
				err := toolCall.Args(&args)
				if err != nil {
					return fmt.Errorf("parse tool call args: %w", err)
				}

				callRes := clientTool.Exec(args.City)
				m := client.Message{
					Role:       "tool",
					Content:    callRes,
					ToolCalls:  nil,
					ToolCallID: toolCall.ID,
				}

				r, err := c.Request(ctx, m)
				if err != nil {
					return fmt.Errorf("requesting llm from processResponse: %w", err)
				}

				err = processResponse(ctx, c, r)
				if err != nil {
					return fmt.Errorf("process response: %w", err)
				}
			}
		}
	}

	fmt.Fprintln(os.Stdout, resp.Content)

	return nil
}
