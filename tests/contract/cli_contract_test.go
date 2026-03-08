package contract_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestCLINoArgsScansCurrentDirectory(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/pdf_watermark_remover")
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected graceful no-arg handling, got error: %v\n%s", err, string(out))
	}
	if !strings.Contains(strings.ToLower(string(out)), "no eligible pdf") {
		t.Fatalf("expected directory scan message, got: %s", string(out))
	}
}

func TestCLIInvalidPathReturnsNonZero(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/pdf_watermark_remover", "does-not-exist.pdf")
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit for invalid path, got success: %s", string(out))
	}
	if !strings.Contains(string(out), "cannot access input file") {
		t.Fatalf("expected actionable error, got: %s", string(out))
	}
}
