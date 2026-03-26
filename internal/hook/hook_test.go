package hook

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	var buf bytes.Buffer
	err := Generate(&buf, "/home/user/.claude/cli-help", false)
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
	if !strings.Contains(script, "${TRY_KEY}.txt") {
		t.Error("expected .txt file lookup pattern with TRY_KEY")
	}

	// Check JSON output format
	if !strings.Contains(script, "hookSpecificOutput") {
		t.Error("missing hookSpecificOutput in JSON output")
	}

	// Check session dedup
	if !strings.Contains(script, "claude-help-cache") {
		t.Error("missing session cache logic")
	}

	// Non-lazy mode should NOT contain lazy collection
	if strings.Contains(script, "lazy_collect") {
		t.Error("non-lazy mode should not contain lazy collection logic")
	}

	// Check subcommand support: LOOKUP_KEY construction
	if !strings.Contains(script, "LOOKUP_KEY") {
		t.Error("expected LOOKUP_KEY for subcommand support")
	}

	// Check longest-match loop
	if !strings.Contains(script, "longest match first") {
		t.Error("expected longest-match lookup comment")
	}
}

func TestGenerateLazy(t *testing.T) {
	var buf bytes.Buffer
	err := Generate(&buf, "/home/user/.claude/cli-help", true)
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

	// Check lazy collection logic
	if !strings.Contains(script, "lazy_collect") {
		t.Error("lazy mode should contain lazy_collect function")
	}

	// Check timeout usage
	if !strings.Contains(script, "timeout 2") {
		t.Error("lazy mode should use timeout for --help collection")
	}

	// Check atomic write
	if !strings.Contains(script, "mktemp") {
		t.Error("lazy mode should use mktemp for atomic writes")
	}

	// Check JSON output format (shared with non-lazy)
	if !strings.Contains(script, "hookSpecificOutput") {
		t.Error("missing hookSpecificOutput in JSON output")
	}

	// Check session dedup (shared with non-lazy)
	if !strings.Contains(script, "claude-help-cache") {
		t.Error("missing session cache logic")
	}

	// Check --lazy comment in header
	if !strings.Contains(script, "--lazy") {
		t.Error("lazy script should mention --lazy in header")
	}

	// Check subcommand support
	if !strings.Contains(script, "LOOKUP_KEY") {
		t.Error("expected LOOKUP_KEY for subcommand support")
	}

	// Check TTL support
	if !strings.Contains(script, "TTL_SECONDS") {
		t.Error("lazy mode should have TTL-based refresh")
	}

	// Check is_expired function
	if !strings.Contains(script, "is_expired") {
		t.Error("lazy mode should have is_expired function")
	}

	// Check help subcommand fallback pattern
	if !strings.Contains(script, "help \"${subs[@]}\"") {
		t.Error("lazy mode should try 'cmd help sub...' fallback")
	}
}
