# CLI Contract: `pdf_watermark_remover`

## Command

```text
pdf_watermark_remover <input-pdf-path>
```

## Inputs

- Positional argument 1: `input-pdf-path`
  - Required
  - Must reference an existing readable PDF file
  - PDF must contain at least 2 pages

## Outputs

- On success:
  - Writes a new PDF in the same directory as input
  - Preferred name: `<input_basename>_output.pdf`
  - Collision policy: if the preferred name exists, write `<input_basename>_output_<n>.pdf` using the next available integer
  - Exit code: `0`

- On failure:
  - Writes no output PDF
  - Prints actionable error message
  - Exit code: non-zero

## Behavioral Guarantees

- Never modifies or overwrites the original input PDF.
- Processes all input pages in original order for successful runs.
- Applies watermark suppression on every page; if watermark consistency is low, uses best-effort suppression and still attempts full-document output.
- Output may be binarized (color preservation is not required).

## Acceptance-oriented checks

- `test/test.pdf` must be processable through this command.
- Output page count must equal input page count.
- Watermark text should not be readable in normal viewing conditions (100%-150% zoom).
