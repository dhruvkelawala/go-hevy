package cmd

import (
	"testing"

	appconfig "github.com/dhruvkelawala/hevy-cli/internal/config"
)

func TestConfigRedact(t *testing.T) {
	got := appconfig.Redact("1234567890abcdef")
	if got != "1234********cdef" {
		t.Fatalf("unexpected redacted key: %s", got)
	}
}

func TestRequirePositivePagination(t *testing.T) {
	if err := requirePositivePagination(0, 5, 10); err == nil {
		t.Fatal("expected page validation error")
	}
	if err := requirePositivePagination(1, 11, 10); err == nil {
		t.Fatal("expected page-size validation error")
	}
	if err := requirePositivePagination(1, 5, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
