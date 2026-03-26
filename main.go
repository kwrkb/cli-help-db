package main

import (
	"os"

	"github.com/kwrkb/cli-help-db/internal/cmd"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:]))
}
