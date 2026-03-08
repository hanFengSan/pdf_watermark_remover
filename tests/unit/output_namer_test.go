package unit_test

import (
	"os"
	"path/filepath"
	"testing"

	"pdf_watermark_remover/internal/output"
)

func TestNextOutputPathIncrements(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "sample.pdf")
	if err := os.WriteFile(in, []byte("stub"), 0o644); err != nil {
		t.Fatal(err)
	}
	first, err := output.NextOutputPath(in)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(first, []byte("stub"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := output.NextOutputPath(in)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("expected incremented output path, got same %s", second)
	}
}
