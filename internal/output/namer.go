package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func NextOutputPath(inputPath string) (string, error) {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	defaultPath := filepath.Join(dir, name+"_remove_watermark.pdf")
	if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
		return defaultPath, nil
	}

	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_remove_watermark_%d.pdf", name, i))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("failed to allocate output path for %s", inputPath)
}
