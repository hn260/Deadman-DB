package main

import (
	"log"

	"deadmandb/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatalf("Error executing CLI: %v", err)
	}
}
