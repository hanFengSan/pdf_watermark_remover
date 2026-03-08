package integration_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	pdfreader "rsc.io/pdf"
)

func testPDFPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "test", "test.pdf")
}

func cleanupOutputs(t *testing.T) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join("..", "..", "test"))
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if (strings.HasPrefix(name, "test_output") || strings.HasPrefix(name, "test_remove_watermark")) && strings.HasSuffix(name, ".pdf") {
			_ = os.Remove(filepath.Join("..", "..", "test", name))
		}
	}
}

func pageCount(t *testing.T, path string) int {
	t.Helper()
	r, err := pdfreader.Open(path)
	if err != nil {
		t.Fatalf("open pdf %s: %v", path, err)
	}
	return r.NumPage()
}
