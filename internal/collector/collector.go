package collector

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Result holds the output of collecting help text for a single command.
type Result struct {
	Name string
	Text string
	Err  error
}

// Executor runs a command and returns its combined output.
// Replaceable for testing.
type Executor func(ctx context.Context, name string, args ...string) (string, error)

// DefaultExecutor runs a real command via exec.CommandContext.
func DefaultExecutor(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	text := string(out)
	// Many CLIs exit non-zero for --help; only fail if no output
	if text == "" && err != nil {
		return "", err
	}
	return text, nil
}

// Options configures collection behavior.
type Options struct {
	Timeout     time.Duration
	LineLimit   int
	Parallelism int
	Exec        Executor
}

// Collect gathers help text for the given commands concurrently.
func Collect(commands []string, opts Options) []Result {
	if opts.Exec == nil {
		opts.Exec = DefaultExecutor
	}
	if opts.Timeout == 0 {
		opts.Timeout = 3 * time.Second
	}
	if opts.LineLimit == 0 {
		opts.LineLimit = 60
	}
	if opts.Parallelism == 0 {
		opts.Parallelism = 8
	}

	results := make([]Result, len(commands))
	sem := make(chan struct{}, opts.Parallelism)
	var wg sync.WaitGroup

	for i, name := range commands {
		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			text, err := collectOne(name, opts)
			results[i] = Result{Name: name, Text: text, Err: err}
		}(i, name)
	}

	wg.Wait()
	return results
}

// splitCommand splits a command name into executable and subcommand args.
// e.g. "docker container ls" -> ("docker", ["container", "ls"])
// e.g. "curl" -> ("curl", [])
func splitCommand(name string) (string, []string) {
	parts := strings.Fields(name)
	if len(parts) <= 1 {
		return name, nil
	}
	return parts[0], parts[1:]
}

func collectOne(name string, opts Options) (string, error) {
	exe, sub := splitCommand(name)

	// Build args safely (avoid mutating sub via append)
	withFlag := func(flag string) []string {
		a := make([]string, len(sub)+1)
		copy(a, sub)
		a[len(sub)] = flag
		return a
	}

	// Try {cmd} {sub...} --help
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	text, err := opts.Exec(ctx, exe, withFlag("--help")...)
	if isUsable(text) {
		return truncate(text, opts.LineLimit), nil
	}

	// Try {cmd} {sub...} -h
	ctx2, cancel2 := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel2()
	text, err = opts.Exec(ctx2, exe, withFlag("-h")...)
	if isUsable(text) {
		return truncate(text, opts.LineLimit), nil
	}

	// Try {cmd} help {sub...} (help subcommand pattern, e.g. "docker help container ls")
	if len(sub) > 0 {
		ctx3, cancel3 := context.WithTimeout(context.Background(), opts.Timeout)
		defer cancel3()
		helpArgs := append([]string{"help"}, sub...)
		text, err = opts.Exec(ctx3, exe, helpArgs...)
		if isUsable(text) {
			return truncate(text, opts.LineLimit), nil
		}
	}

	// Try man: for subcommands use hyphenated form (e.g. "man git-remote")
	ctx4, cancel4 := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel4()
	manPage := strings.Join(append([]string{exe}, sub...), "-")
	text, err = opts.Exec(ctx4, "man", manPage)
	if isUsable(text) {
		text = stripManFormatting(text)
		return truncate(text, opts.LineLimit), nil
	}

	if err != nil {
		return "", err
	}
	return "", nil
}

func isUsable(text string) bool {
	return len(strings.TrimSpace(text)) >= 10
}

func truncate(text string, maxLines int) string {
	lines := strings.SplitN(text, "\n", maxLines+1)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}

func stripManFormatting(text string) string {
	// Equivalent to `col -b`: remove backspace sequences used for bold/underline
	var buf bytes.Buffer
	for i := 0; i < len(text); i++ {
		if i+1 < len(text) && text[i+1] == '\b' {
			i++ // skip the char and the backspace; next iteration picks the overprint
			continue
		}
		if text[i] == '\b' {
			// Remove previous character from buf if possible
			if buf.Len() > 0 {
				b := buf.Bytes()
				buf.Reset()
				buf.Write(b[:len(b)-1])
			}
			continue
		}
		buf.WriteByte(text[i])
	}
	return buf.String()
}
