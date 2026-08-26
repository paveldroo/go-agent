package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/paveldroo/go-agent/request"
)

func main() {
	var task string
	flag.StringVar(&task, "task", "", "task for agent")

	flag.Parse()

	if len(task) == 0 {
		fmt.Fprint(os.Stderr, "you should define a task for agent\n")
		os.Exit(1)
	}

	client := request.New()

	resp, err := client.Request(task)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error requesting llm: %s\n", err.Error())
	}

	fmt.Fprintln(os.Stdout, resp)
}
