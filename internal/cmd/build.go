package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/kwrkb/cli-help-db/internal/collector"
	"github.com/kwrkb/cli-help-db/internal/config"
	"github.com/kwrkb/cli-help-db/internal/db"
	"github.com/kwrkb/cli-help-db/internal/scanner"
)

func runBuild(args []string) int {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to config file")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	commands := cfg.Commands
	if len(commands) == 0 {
		fmt.Fprintln(os.Stderr, "error: no commands configured. Add commands to config file or pass them as arguments.")
		return 1
	}

	// Filter to commands that actually exist on PATH
	commands = scanner.Filter(commands)
	if len(commands) == 0 {
		fmt.Fprintln(os.Stderr, "error: none of the configured commands were found on $PATH")
		return 1
	}

	fmt.Fprintf(os.Stderr, "collecting help for %d commands...\n", len(commands))

	results := collector.Collect(commands, collector.Options{
		Timeout:     cfg.Timeout,
		LineLimit:   cfg.LineLimit,
		Parallelism: cfg.Parallelism,
	})

	store := db.New(cfg.OutputDir)
	var succeeded, failed, skipped int
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "  FAIL  %s: %v\n", r.Name, r.Err)
			failed++
			continue
		}
		if r.Text == "" {
			fmt.Fprintf(os.Stderr, "  SKIP  %s: no help text\n", r.Name)
			skipped++
			continue
		}
		if err := store.Write(r.Name, r.Text); err != nil {
			fmt.Fprintf(os.Stderr, "  FAIL  %s: %v\n", r.Name, err)
			failed++
			continue
		}
		fmt.Fprintf(os.Stderr, "  OK    %s\n", r.Name)
		succeeded++
	}

	fmt.Fprintf(os.Stderr, "\ndone: %d succeeded, %d failed, %d skipped\n", succeeded, failed, skipped)
	fmt.Fprintf(os.Stderr, "output: %s\n", cfg.OutputDir)

	if failed > 0 {
		return 1
	}
	return 0
}

func loadConfig(path string) (*config.Config, error) {
	if path != "" {
		return config.LoadFrom(path)
	}
	return config.Load()
}
