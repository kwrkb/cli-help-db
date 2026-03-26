package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/kwrkb/cli-help-db/internal/scanner"
)

func runScan(args []string) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	names := scanner.Scan()
	for _, name := range names {
		fmt.Println(name)
	}
	fmt.Fprintf(os.Stderr, "\n%d commands found\n", len(names))
	return 0
}
