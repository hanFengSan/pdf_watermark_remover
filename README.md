# PDF Watermark Remover

A local CLI tool that removes repeated text-like watermarks from PDF files by rendering pages to images, applying watermark suppression heuristics, and rebuilding a cleaned PDF.

## Highlights

- No external Poppler dependency (`pdftoppm` removed).
- Uses `go-pdfium` in WebAssembly mode (no CGO required).
- Cross-platform build support (including Windows `.exe`).
- Supports single-file mode and batch mode.
- Batch mode excludes files containing `remove_watermark` in filename.
- Progress logs include per-file prefixes, e.g. `[1/10 sample.pdf] ...`.

## How It Works

1. Validate input PDF.
2. Render each page to PNG (400 DPI by default).
3. Estimate watermark pattern and page suppression strategy.
4. Suppress watermark + binarize + readability cleanup.
5. Retry selected pages if residual artifacts are detected.
6. Rebuild final PDF from processed images.

## Build

Requirements:

- Go 1.25+

Build current platform binary:

```bash
go build -o pdf_watermark_remover ./cmd/pdf_watermark_remover
```

Build Windows amd64 executable:

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/pdf_watermark_remover_windows_amd64.exe ./cmd/pdf_watermark_remover
```

## Usage

### 1) Single file mode

```bash
./pdf_watermark_remover /path/to/input.pdf
```

Output file naming:

- `input_remove_watermark.pdf`
- If exists: `input_remove_watermark_1.pdf`, `input_remove_watermark_2.pdf`, ...

### 2) Batch mode (no arguments)

```bash
./pdf_watermark_remover
```

Behavior:

- Scans the executable directory (non-recursive) for `*.pdf`.
- Excludes files containing `remove_watermark` in filename.
- Prompts for confirmation (`y/yes` to proceed).
- Processes files sequentially.
- Prints total elapsed time after all files are done.
- Waits for Enter before exit.

## Environment Variables

Watermark / debug:

- `WM_MODE=single|hybrid` (default: `single`)
- `WM_DEBUG_PAGE=1` (prints fallback diagnostics)
- `WM_DEBUG_MASK=1` (writes debug masks to temp work dir)

Render parallelism:

- `WM_RENDER_WORKERS=<n>`
  - Explicit render worker count (clamped to CPU cores).
- `WM_RENDER_CPU_TARGET=<10-100>`
  - Used when `WM_RENDER_WORKERS` is not set.

Process parallelism:

- `WM_PROCESS_WORKER=<n>`
  - Page processing worker count.
  - If not set, defaults to `4` on machines with >= 4 cores.

## Tests

Run all tests:

```bash
go test ./...
```

Focused suites:

```bash
go test ./tests/contract -run TestCLI
go test ./tests/integration -run TestUS1
go test ./tests/integration -run TestUS2
go test ./tests/integration -run TestUS3
go test ./tests/unit
```

## Notes

- This is an image-based pipeline; vector/text semantics are not preserved.
- Quality depends on watermark style and heuristics.
- For large PDFs, tune worker counts to balance speed and memory usage.

## Project Homepage

https://github.com/hanFengSan/pdf_watermark_remover
