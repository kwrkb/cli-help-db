package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/kwrkb/cli-help-db/internal/config"
	"github.com/kwrkb/cli-help-db/internal/hook"
)

func runHook(args []string) int {
	fs := flag.NewFlagSet("hook", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to config file")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	if err := hook.Generate(os.Stdout, cfg.OutputDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}
