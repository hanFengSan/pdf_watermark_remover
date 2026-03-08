package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pdfreader "rsc.io/pdf"
)

type InputInfo struct {
	Path      string
	PageCount int
}

func ValidateInput(path string) (InputInfo, error) {
	if path == "" {
		return InputInfo{}, &ExitError{Code: ExitInvalidInput, Msg: "input path is required"}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return InputInfo{}, &ExitError{Code: ExitInvalidInput, Msg: fmt.Sprintf("invalid path: %v", err)}
	}
	st, err := os.Stat(abs)
	if err != nil {
		return InputInfo{}, &ExitError{Code: ExitInvalidInput, Msg: fmt.Sprintf("cannot access input file: %v", err)}
	}
	if st.IsDir() {
		return InputInfo{}, &ExitError{Code: ExitInvalidInput, Msg: "input path points to a directory"}
	}
	if strings.ToLower(filepath.Ext(abs)) != ".pdf" {
		return InputInfo{}, &ExitError{Code: ExitInvalidInput, Msg: "input file must be a PDF"}
	}

	r, err := pdfreader.Open(abs)
	if err != nil {
		return InputInfo{}, &ExitError{Code: ExitInvalidInput, Msg: fmt.Sprintf("failed to parse PDF: %v", err)}
	}
	pages := r.NumPage()
	if pages < 2 {
		return InputInfo{}, &ExitError{Code: ExitInvalidInput, Msg: "input PDF must contain at least 2 pages"}
	}

	return InputInfo{Path: abs, PageCount: pages}, nil
}
