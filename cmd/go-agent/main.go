package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/paveldroo/go-agent/client"
	"github.com/paveldroo/go-agent/config"
)

var errDefineTask = errors.New("you should define a task for agent")

func main() {
	var task string
	flag.StringVar(&task, "task", "", "task for agent")

	flag.Parse()

	err := run(task)
	if err != nil {
		fmt.Fprintf(os.Stderr, "go-agent: error: %s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func run(task string) error {
	if len(task) == 0 {
		return errDefineTask
	}

	cfg := config.New()
	c := client.New(cfg)

	resp, err := c.Request(task)
	if err != nil {
		return fmt.Errorf("requesting llm: %w", err)
	}

	fmt.Fprintln(os.Stdout, resp)

	return nil
}
