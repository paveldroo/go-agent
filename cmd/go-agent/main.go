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
	c := client.New(cfg)

	ctx := context.Background()

	resp, err := c.Request(ctx, prompt)
	if err != nil {
		return fmt.Errorf("requesting llm: %w", err)
	}

	err = printResult(resp)
	if err != nil {
		return fmt.Errorf("print result: %w", err)
	}

	return nil
}

func printResult(resp *client.LLMResponse) error {
	content := resp.Content
	if resp.FinishReason == client.ReasonToolCalls {
		if len(resp.ToolCalls) == 0 {
			return fmt.Errorf("%w: no tool calls in response", client.ErrToolCallsCorrupted)
		}
		firstToolCall := resp.ToolCalls[0]

		weatherArgs := tool.WeatherArgs{
			City: "",
		}
		err := firstToolCall.Args(&weatherArgs)
		if err != nil {
			return fmt.Errorf("unmarshal tool call args: %w", err)
		}

		content = fmt.Sprintf("%s(city=%q)", firstToolCall.Function.Name, weatherArgs.City)
	}

	fmt.Fprintln(os.Stdout, content)

	return nil
}
