# pdf_watermarker_remover Development Guidelines

Last updated: 2026-03-08

## Active Technologies

- Go 1.25+ (`go 1.25.0`, `toolchain go1.26.0`)
- `github.com/klippa-app/go-pdfium v1.18.0` (WebAssembly mode via wazero, no CGO required)
- `github.com/phpdave11/gofpdf v1.4.2` (image pages -> output PDF rebuild)
- `rsc.io/pdf v0.1.1` (input validation/page count checks)

## Current Processing Architecture

1. Validate input PDF (`internal/pipeline/input_validator.go`)
2. Render each PDF page to PNG at 400 DPI (`internal/pdf/render.go`)
3. Estimate watermark pattern and per-page mode (`internal/watermark/*.go`)
4. Suppress watermark and binarize (`internal/watermark/suppressor.go`, `internal/imageproc/binarize.go`)
5. Apply readability cleanup and artifact-based retry fallback (`internal/imageproc/readability_guard.go`, `internal/imageproc/artifact_score.go`)
6. Rebuild final PDF from processed PNG pages (`internal/pdf/rebuild.go`)

## Repository Layout

```text
cmd/pdf_watermark_remover/main.go       # CLI entrypoint
internal/pipeline/                      # Orchestration, worker pool, tuning, errors
internal/pdf/                           # PDF render/rebuild adapters
internal/watermark/                     # Watermark pattern estimation + suppression
internal/imageproc/                     # Binarization, readability guard, component analysis
internal/output/                        # Output naming strategy
tests/contract/                         # CLI contract tests
tests/integration/                      # End-to-end behavior tests
tests/unit/                             # Unit tests for core logic
test/test.pdf                           # Main fixture
```

## Build and Test Commands

- Build local binary:
  - `go build -o pdf_watermark_remover ./cmd/pdf_watermark_remover`
- Run all tests:
  - `go test ./...`
- Focused tests:
  - `go test ./tests/contract -run TestCLI`
  - `go test ./tests/integration -run TestUS1`
  - `go test ./tests/integration -run TestUS2`
  - `go test ./tests/integration -run TestUS3`
  - `go test ./tests/unit`
- Windows cross-compile:
  - `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/pdf_watermark_remover_windows_amd64.exe ./cmd/pdf_watermark_remover`
  - `GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -o dist/pdf_watermark_remover_windows_arm64.exe ./cmd/pdf_watermark_remover`

## Runtime Environment Flags

- `WM_MODE=single|hybrid`
  - `single` is default.
  - `hybrid` uses template-informed suppression path.
- `WM_DEBUG_PAGE=1`
  - Prints per-page fallback diagnostics (artifact score/count, retry decision).
- `WM_DEBUG_MASK=1`
  - Saves per-page watermark masks into temp work dir for debugging.
- `WM_RENDER_WORKERS=<n>`
  - Overrides PDF render worker count.
  - Clamped to available CPU cores.
- `WM_RENDER_CPU_TARGET=<10-100>`
  - Target CPU usage percentage used to derive render worker count when `WM_RENDER_WORKERS` is not set.
  - Default is `80`.

## Tuning and Heuristic Controls

- Centralized tuning config: `internal/pipeline/tuning.go`
- If quality changes are needed, adjust values there first before editing algorithm internals.
- Current retry/fallback behavior:
  - Fallback triggers when residual artifact score is high and fragment count threshold is met.
  - Retry uses more aggressive suppression+binarization thresholds.

## Maintainability Rules

- Keep orchestration in `pipeline`, not in `cmd`.
- Keep heuristics/thresholds centralized in `TuningConfig`; avoid new scattered magic numbers.
- Reuse shared connected-component utilities in `internal/imageproc/components.go`.
- Prefer adding unit tests for algorithm changes and integration tests for pipeline behavior.
- Preserve deterministic output naming behavior in `internal/output/namer.go`.

## Windows Release Notes

- Current renderer uses go-pdfium WebAssembly mode, so no external Poppler/pdftoppm install is required.
- Binary is self-contained from app dependency perspective (no CGO/DLL runtime requirement introduced by this project code).
- Validate on real Win10/Win11 machines for performance and memory before release.

## Recent Changes

- Replaced external `pdftoppm` rendering dependency with `go-pdfium` WebAssembly renderer in `internal/pdf/render.go`.
- Upgraded project baseline to Go 1.25+ and `go-pdfium v1.18.0`.
- Refactored pipeline maintainability:
  - Added centralized tuning config (`internal/pipeline/tuning.go`)
  - Extracted page processing/fallback helper functions from `runner.go`
  - Introduced shared connected-component analysis (`internal/imageproc/components.go`)
  - Removed duplicate component-scanning logic from readability/artifact modules.

## Known Constraints

- Processing is currently image-based; text/vector semantics are not preserved.
- Output pages may be binarized to maximize watermark suppression consistency.
- Heuristic-based suppression may require tuning for new watermark styles.
