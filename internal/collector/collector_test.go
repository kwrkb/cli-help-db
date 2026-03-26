package collector

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func fakeExec(responses map[string]string) Executor {
	return func(ctx context.Context, name string, args ...string) (string, error) {
		key := name
		if len(args) > 0 {
			key = name + " " + strings.Join(args, " ")
		}
		if text, ok := responses[key]; ok {
			return text, nil
		}
		return "", fmt.Errorf("command not found: %s", key)
	}
}

func TestCollect_HelpFallback(t *testing.T) {
	exec := fakeExec(map[string]string{
		"mycmd --help": "", // empty — should fall back
		"mycmd -h":     "Usage: mycmd [options]\n\nA useful tool for things.",
	})

	results := Collect([]string{"mycmd"}, Options{Exec: exec})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if !strings.Contains(results[0].Text, "Usage: mycmd") {
		t.Errorf("expected help text, got %q", results[0].Text)
	}
}

func TestCollect_ManFallback(t *testing.T) {
	exec := fakeExec(map[string]string{
		"mycmd --help": "",
		"mycmd -h":     "",
		"man mycmd":    "NAME\n     mycmd - does things\n\nSYNOPSIS\n     mycmd [options]\n\nDESCRIPTION\n     Long description here.",
	})

	results := Collect([]string{"mycmd"}, Options{Exec: exec})
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if !strings.Contains(results[0].Text, "mycmd") {
		t.Errorf("expected man page content, got %q", results[0].Text)
	}
}

func TestTruncate(t *testing.T) {
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	text := strings.Join(lines, "\n")

	got := truncate(text, 5)
	gotLines := strings.Split(got, "\n")
	if len(gotLines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(gotLines))
	}
	if gotLines[0] != "line 1" {
		t.Errorf("first line = %q, want 'line 1'", gotLines[0])
	}
}

func TestCollect_Parallel(t *testing.T) {
	exec := fakeExec(map[string]string{
		"a --help": "Usage: a\n\nCommand a does stuff.",
		"b --help": "Usage: b\n\nCommand b does stuff.",
		"c --help": "Usage: c\n\nCommand c does stuff.",
	})

	results := Collect([]string{"a", "b", "c"}, Options{Exec: exec, Parallelism: 2})
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.Name, r.Err)
		}
		if r.Text == "" {
			t.Errorf("empty text for %s", r.Name)
		}
	}
}

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name    string
		wantExe string
		wantSub []string
	}{
		{"curl", "curl", nil},
		{"docker container ls", "docker", []string{"container", "ls"}},
		{"kubectl get", "kubectl", []string{"get"}},
	}
	for _, tt := range tests {
		exe, sub := splitCommand(tt.name)
		if exe != tt.wantExe {
			t.Errorf("splitCommand(%q) exe = %q, want %q", tt.name, exe, tt.wantExe)
		}
		if len(sub) != len(tt.wantSub) {
			t.Errorf("splitCommand(%q) sub = %v, want %v", tt.name, sub, tt.wantSub)
			continue
		}
		for i := range sub {
			if sub[i] != tt.wantSub[i] {
				t.Errorf("splitCommand(%q) sub[%d] = %q, want %q", tt.name, i, sub[i], tt.wantSub[i])
			}
		}
	}
}

func TestCollect_Subcommand(t *testing.T) {
	exec := fakeExec(map[string]string{
		"docker container ls --help": "Usage: docker container ls [OPTIONS]\n\nList containers\n\nOptions:\n  -a, --all  Show all",
	})

	results := Collect([]string{"docker container ls"}, Options{Exec: exec})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if !strings.Contains(results[0].Text, "docker container ls") {
		t.Errorf("expected subcommand help, got %q", results[0].Text)
	}
}

func TestCollect_SubcommandHelpFallback(t *testing.T) {
	// --help and -h fail, but "docker help container ls" succeeds
	exec := fakeExec(map[string]string{
		"docker container ls --help": "",
		"docker container ls -h":     "",
		"docker help container ls":   "Usage: docker container ls [OPTIONS]\n\nList containers",
	})

	results := Collect([]string{"docker container ls"}, Options{Exec: exec})
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if !strings.Contains(results[0].Text, "docker container ls") {
		t.Errorf("expected help subcommand fallback, got %q", results[0].Text)
	}
}

func TestStripManFormatting(t *testing.T) {
	// Bold: char + backspace + char
	input := "H\bHe\bel\bll\blo\bo"
	got := stripManFormatting(input)
	if got != "Hello" {
		t.Errorf("stripManFormatting(%q) = %q, want 'Hello'", input, got)
	}
}
