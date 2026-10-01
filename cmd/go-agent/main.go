package main

import (
	"errors"
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
	conv := conversation.New()
	err = conv.Run(c, prompt)
	if err != nil {
		return fmt.Errorf("run conversation: %w", err)
	}

	return nil
}
