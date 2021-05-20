package main

import (
	"github.com/bsc-bridge-svc/internal/cli"
	"os"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
