package main

import (
	"flag"
	"log"
)

func main() {
	mode := flag.String("mode", "cli", "Run mode")
	task := flag.String("task", "", "Task instruction")
    flag.Parse()

	switch *mode {
    case "cli":
		runCli(*task)

    case "lark":
		runLark()

	default:
		log.Fatalf("unsupported mode: %s", *mode)
	}
}
