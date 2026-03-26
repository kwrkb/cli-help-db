package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/kwrkb/cli-help-db/internal/config"
	"github.com/kwrkb/cli-help-db/internal/db"
)

func runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to config file")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	store := db.New(cfg.OutputDir)
	names, err := store.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if len(names) == 0 {
		fmt.Fprintln(os.Stderr, "database is empty")
		return 0
	}

	for _, name := range names {
		fmt.Println(name)
	}
	fmt.Fprintf(os.Stderr, "\n%d commands in database\n", len(names))
	return 0
}
