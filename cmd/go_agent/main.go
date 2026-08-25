package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var task string
	flag.StringVar(&task, "task", "", "task for agent")

	flag.Parse()

	if len(task) == 0 {
		fmt.Fprint(os.Stderr, "you should define a task for agent\n")
		os.Exit(1)
	}

	fmt.Println(task)
}
