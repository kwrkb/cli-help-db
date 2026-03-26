package cmd

import (
	"fmt"
	"os"
)

const usage = `Usage: cli-help-db <command> [options]

Commands:
  scan     List executable commands on $PATH
  build    Collect --help output and build the database
  list     Show commands stored in the database
  hook     Generate auto-help.sh hook script

Options:
  -h, --help    Show this help
`

func Run(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return 1
	}

	switch args[0] {
	case "scan":
		return runScan(args[1:])
	case "build":
		return runBuild(args[1:])
	case "list":
		return runList(args[1:])
	case "hook":
		return runHook(args[1:])
	case "-h", "--help", "help":
		fmt.Print(usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", args[0], usage)
		return 1
	}
}
