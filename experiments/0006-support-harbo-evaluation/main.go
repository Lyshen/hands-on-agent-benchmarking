package main

import (
	"flag"
	"log"
)

func main() {
	mode := flag.String("mode", "cli", "Run mode")
    flag.Parse()

	switch *mode {
    case "cli":
		runCli()

    case "lark":
		runLark()

	default:
		log.Fatalf("unsupported mode: %s", *mode)
	}
}
