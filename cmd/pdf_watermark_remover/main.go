package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"pdf_watermark_remover/internal/logutil"
	"pdf_watermark_remover/internal/pipeline"
)

func main() {
	fmt.Println("project homepage: https://github.com/hanFengSan/pdf_watermark_remover")

	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [input-pdf-path]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "note: output pages may be binarized to improve watermark suppression readability")
		fmt.Fprintln(os.Stderr, "note: WM_MODE=single (default) or WM_MODE=hybrid")
		waitForAnyKeyAndExit(pipeline.ExitInvalidInput)
	}

	inputs := make([]string, 0, 1)
	if len(os.Args) == 2 {
		inputs = append(inputs, os.Args[1])
	} else {
		files, err := discoverBatchInputs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to scan current directory: %v\n", err)
			waitForAnyKeyAndExit(pipeline.ExitInvalidInput)
		}
		if len(files) == 0 {
			fmt.Println("no eligible PDF files found in current directory")
			waitForAnyKeyAndExit(0)
		}

		fmt.Printf("found %d eligible PDF file(s):\n", len(files))
		for i, f := range files {
			fmt.Printf("%d) %s\n", i+1, filepath.Base(f))
		}
		fmt.Print("process all files? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			fmt.Println("batch processing cancelled")
			waitForAnyKeyAndExit(0)
		}
		inputs = files
	}

	runner := pipeline.NewRunner()
	started := time.Now()

	for i, inputPath := range inputs {
		prefix := fmt.Sprintf("[%d/%d %s]", i+1, len(inputs), filepath.Base(inputPath))
		_ = os.Setenv("WM_LOG_PREFIX", prefix)

		result, err := runner.Run(context.Background(), inputPath)
		if err != nil {
			if ee, ok := pipeline.AsExitError(err); ok {
				fmt.Fprintf(os.Stderr, "%s %s\n", prefix, ee.Error())
				waitForAnyKeyAndExit(ee.Code)
			}
			fmt.Fprintf(os.Stderr, "%s processing failed: %v\n", prefix, err)
			waitForAnyKeyAndExit(pipeline.ExitProcessing)
		}

		logutil.Printf("input: %s\n", result.InputPath)
		logutil.Printf("output: %s\n", result.OutputPath)
		logutil.Printf("pages: %d\n", result.PageCount)
		logutil.Printf("duration: %s\n", result.Duration)
	}
	_ = os.Unsetenv("WM_LOG_PREFIX")

	fmt.Printf("all tasks completed in %s\n", time.Since(started))
	waitForAnyKeyAndExit(0)
}

func discoverBatchInputs() ([]string, error) {
	scanDir, err := executableDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(scanDir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".pdf") {
			continue
		}
		if strings.Contains(lower, "remove_watermark") {
			continue
		}
		files = append(files, filepath.Join(scanDir, name))
	}
	sort.Strings(files)
	return files, nil
}

func executableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	realPath, err := filepath.EvalSymlinks(exePath)
	if err == nil {
		exePath = realPath
	}
	return filepath.Dir(exePath), nil
}

func waitForAnyKeyAndExit(code int) {
	fmt.Print("press Enter to exit...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	os.Exit(code)
}
