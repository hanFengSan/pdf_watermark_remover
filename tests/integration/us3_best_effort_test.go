package integration_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRootUS3BestEffort(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestUS3BestEffortContinuesProcessing(t *testing.T) {
	cleanupOutputs(t)
	in := testPDFPath(t)
	cmd := exec.Command("go", "run", "./cmd/pdf_watermark_remover", in)
	cmd.Dir = repoRootUS3BestEffort(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected successful best-effort processing: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "output:") {
		t.Fatalf("expected output summary in CLI output, got: %s", string(out))
	}
}
