package redact

import (
	"os"
	"strings"
	"testing"
)

func TestRedactCommand_ModeOff_NoChange(t *testing.T) {
	cmd := "echo hello"

	// Default redact mode is ModeOff when REWIND_REDACT is unset or unknown.
	// Make the test hermetic.
	_ = os.Unsetenv("REWIND_REDACT")

	got := RedactCommand(cmd)
	if got != cmd {
		t.Fatalf("RedactCommand(%q) = %q, want %q", cmd, got, cmd)
	}
}

func TestRedactCommand_ModeSkip_SkipsWhenSecretFound(t *testing.T) {
	cmd := "gh auth login --with-token ghp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	_ = os.Setenv("REWIND_REDACT", "skip")

	got := RedactCommand(cmd)
	if got != "" {
		t.Fatalf("RedactCommand(%q) = %q, want empty string", cmd, got)
	}
}

func TestRedactCommand_ModeRedact_RedactsPAT(t *testing.T) {
	cmd := "gh auth login --with-token ghp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	_ = os.Setenv("REWIND_REDACT", "redact")

	got := RedactCommand(cmd)
	if got == "" {
		t.Fatalf("RedactCommand(%q) returned empty", cmd)
	}
	if got == cmd {
		t.Fatalf("RedactCommand(%q) did not change", cmd)
	}
	if strings.Contains(got, "ghp_") {
		t.Fatalf("Redacted command still contains ghp_: %q", got)
	}
}

