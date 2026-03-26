package hook

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	var buf bytes.Buffer
	err := Generate(&buf, "/home/user/.claude/cli-help")
	if err != nil {
		t.Fatal(err)
	}

	script := buf.String()

	// Check shebang
	if !strings.HasPrefix(script, "#!/usr/bin/env bash") {
		t.Error("missing shebang")
	}

	// Check DB dir is embedded
	if !strings.Contains(script, "/home/user/.claude/cli-help") {
		t.Error("DB directory not found in generated script")
	}

	// Check it reads from .txt files
	if !strings.Contains(script, "${BASE_CMD}.txt") {
		t.Error("expected .txt file lookup pattern")
	}

	// Check JSON output format
	if !strings.Contains(script, "hookSpecificOutput") {
		t.Error("missing hookSpecificOutput in JSON output")
	}

	// Check session dedup
	if !strings.Contains(script, "claude-help-cache") {
		t.Error("missing session cache logic")
	}
}
