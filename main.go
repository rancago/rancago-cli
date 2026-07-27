package main

import (
	"os"

	"github.com/rancago/rancago-cli/commands"
)

func main() {
	os.Exit(commands.RunCLI())
}
