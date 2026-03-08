# Quickstart: PDF Watermark Removal CLI

## Prerequisites

- Go 1.22+
- Poppler tools available in PATH (`pdftoppm`)
- Repository checked out on branch `001-remove-pdf-watermark`

## Build

```bash
go build -o pdf_watermark_remover ./cmd/pdf_watermark_remover
```

## Run

```bash
./pdf_watermark_remover test/test.pdf
```

Mode selection (single-page analysis is default):

```bash
WM_MODE=single ./pdf_watermark_remover test/test.pdf
WM_MODE=hybrid ./pdf_watermark_remover test/test.pdf
```

Expected behavior:
- Creates `test/test_output.pdf` if not present.
- If `test/test_output.pdf` already exists, creates `test/test_output_1.pdf` (or next available increment).
- Does not modify `test/test.pdf`.

## Verify output

1. Confirm output PDF exists in `test/`.
2. Open output and visually confirm watermark text is not readable at 100%-150% zoom.
3. Confirm output page count equals input page count.

## Run tests

```bash
go test ./...
```

For focused story checks:

```bash
go test ./tests/contract -run TestCLI
go test ./tests/integration -run TestUS1
go test ./tests/integration -run TestUS2
go test ./tests/integration -run TestUS3
```

## Performance check

- Use an eligible ~100-page document and verify processing completes in <=5 minutes under normal local conditions.
- For the provided fixture, run `time ./pdf_watermark_remover test/test.pdf` and confirm successful completion and output page-count parity.
