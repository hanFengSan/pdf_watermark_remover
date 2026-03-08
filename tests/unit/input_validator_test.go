package unit_test

import (
	"os"
	"path/filepath"
	"testing"

	"pdf_watermark_remover/internal/pipeline"
)

func TestValidateInputRejectsNonPDF(t *testing.T) {
	f := filepath.Join(t.TempDir(), "not-pdf.txt")
	if err := os.WriteFile(f, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := pipeline.ValidateInput(f); err == nil {
		t.Fatalf("expected non-pdf validation error")
	}
}
