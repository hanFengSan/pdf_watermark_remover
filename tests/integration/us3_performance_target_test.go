package integration_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func repoRootUS3Perf(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestUS3PerformanceTarget(t *testing.T) {
	cleanupOutputs(t)
	in := testPDFPath(t)
	start := time.Now()
	cmd := exec.Command("go", "run", "./cmd/pdf_watermark_remover", in)
	cmd.Dir = repoRootUS3Perf(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("processing failed: %v\n%s", err, string(out))
	}
	if d := time.Since(start); d > 5*time.Minute {
		t.Fatalf("processing exceeded 5 minute target: %s", d)
	}
}
