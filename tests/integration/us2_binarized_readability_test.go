package integration_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRootUS2(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestUS2BinarizedOutputSignal(t *testing.T) {
	cleanupOutputs(t)
	in := testPDFPath(t)
	cmd := exec.Command("go", "run", "./cmd/pdf_watermark_remover", in)
	cmd.Dir = repoRootUS2(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("processing failed: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "output:") {
		t.Fatalf("expected output path in CLI output, got: %s", string(out))
	}
}
