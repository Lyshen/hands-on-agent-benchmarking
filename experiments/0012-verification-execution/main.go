package main

import (
	"flag"
	"log"
)

func main() {
	mode := flag.String("mode", "cli", "Run mode")
	task := flag.String("task", "", "Task instruction")
	config := flag.String("config", "conf/config.local.json", "The directory of conf file")
	flag.Parse()

	switch *mode {
	case "cli":
		runCli(*task, *config)

	case "lark":
		runLark(*config)

	default:
		log.Fatalf("unsupported mode: %s", *mode)
	}
}
