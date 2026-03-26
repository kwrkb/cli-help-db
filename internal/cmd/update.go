package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/kwrkb/cli-help-db/internal/collector"
	"github.com/kwrkb/cli-help-db/internal/db"
	"github.com/kwrkb/cli-help-db/internal/scanner"
)

func runUpdate(args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to config file")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if len(cfg.Commands) == 0 {
		fmt.Fprintln(os.Stderr, "error: no commands configured")
		return 1
	}

	store := db.New(cfg.OutputDir)

	// Find commands not yet in DB
	var missing []string
	for _, name := range cfg.Commands {
		if !store.Has(name) {
			missing = append(missing, name)
		}
	}

	if len(missing) == 0 {
		fmt.Fprintln(os.Stderr, "database is up to date")
		return 0
	}

	// Filter to commands that exist on PATH
	missing = scanner.Filter(missing)
	if len(missing) == 0 {
		fmt.Fprintln(os.Stderr, "no new commands found on $PATH")
		return 0
	}

	fmt.Fprintf(os.Stderr, "updating %d new commands...\n", len(missing))

	results := collector.Collect(missing, collector.Options{
		Timeout:     cfg.Timeout,
		LineLimit:   cfg.LineLimit,
		Parallelism: cfg.Parallelism,
	})

	var succeeded, failed int
	for _, r := range results {
		if r.Err != nil || r.Text == "" {
			failed++
			continue
		}
		if err := store.Write(r.Name, r.Text); err != nil {
			failed++
			continue
		}
		fmt.Fprintf(os.Stderr, "  OK    %s\n", r.Name)
		succeeded++
	}

	fmt.Fprintf(os.Stderr, "\ndone: %d added, %d failed\n", succeeded, failed)
	return 0
}
