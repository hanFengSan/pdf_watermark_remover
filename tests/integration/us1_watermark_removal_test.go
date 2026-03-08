package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestUS1ProcessFixtureAndPreservePageCount(t *testing.T) {
	cleanupOutputs(t)
	in := testPDFPath(t)
	cmd := exec.Command("go", "run", "./cmd/pdf_watermark_remover", in)
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("processing failed: %v\n%s", err, string(out))
	}

	outPath := filepath.Join("..", "..", "test", "test_remove_watermark.pdf")
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output PDF at %s", outPath)
	}

	inPages := pageCount(t, in)
	outPages := pageCount(t, outPath)
	if inPages != outPages {
		t.Fatalf("page count mismatch: in=%d out=%d", inPages, outPages)
	}
}
