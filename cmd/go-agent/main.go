package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/paveldroo/go-agent/client"
	"github.com/paveldroo/go-agent/config"
)

func main() {
	var task string
	flag.StringVar(&task, "task", "", "task for agent")

	flag.Parse()

	if len(task) == 0 {
		fmt.Fprint(os.Stderr, "you should define a task for agent\n")
		os.Exit(1)
	}

	cfg := config.New()
	c := client.New(cfg)

	resp, err := c.Request(task)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error requesting llm: %s\n", err.Error())
	}

	fmt.Fprintln(os.Stdout, resp)
}
