package cleaner

import "testing"

func TestCleanANSI_RemovesANSI(t *testing.T) {
	input := "\033[31mred\033[0m"
	got := CleanANSI(input)
	if got != "red" {
		t.Fatalf("CleanANSI(%q) = %q, want %q", input, got, "red")
	}
}

func TestCleanANSI_RemovesSpinnerCharacters(t *testing.T) {
	input := "Loading \u2801\u2802\u2803"
	got := CleanANSI(input)
	if got != "Loading" {
		t.Fatalf("CleanANSI(%q) = %q, want %q", input, got, "Loading")
	}
}

func TestCleanANSI_NoANSIUnchanged(t *testing.T) {
	input := "hello world"
	got := CleanANSI(input)
	if got != input {
		t.Fatalf("CleanANSI(%q) = %q, want %q", input, got, input)
	}
}

func TestCleanANSI_EmptyDoesNotPanic(t *testing.T) {
	got := CleanANSI("")
	if got != "" {
		t.Fatalf("CleanANSI("") = %q, want %q", got, "")
	}
}

