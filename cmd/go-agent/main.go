package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/paveldroo/go-agent/client"
	"github.com/paveldroo/go-agent/config"
	"github.com/paveldroo/go-agent/conversation"
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
	maxStepsPtr := flag.Int("max-steps", 0, "Max steps for agent")
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 { //nolint:mnd // convenient args usage
		return errMissedArgument
	}

	prompt := args[0]

	if len(prompt) == 0 {
		return errDefineTask
	}

	cfg, err := config.New(*maxStepsPtr)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	ctx := context.Background()

	conv := conversation.New()

	c := client.New(cfg, tool.WeatherTool(), tool.CapitalTool(), tool.InfiniteTool())
	err = conv.Run(ctx, c, prompt)
	if err != nil {
		return fmt.Errorf("run conversation: %w", err)
	}

	return nil
}
