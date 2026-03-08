# Phase 0 Research: PDF Watermark Removal CLI

## Decision 1: PDF page rasterization and rebuild pipeline

- Decision: Use external `pdftoppm` (Poppler) for PDF-to-image rasterization, then rebuild PDF from processed page images with `pdfcpu`.
- Rationale: This gives stable rendering quality for scanned/image PDFs, predictable behavior on multi-page documents, and keeps Go application logic focused on processing instead of low-level rendering internals.
- Alternatives considered:
  - MuPDF (`mutool draw`): often faster, but introduces additional licensing/distribution considerations.
  - `go-pdfium`: good in-process control, but increases native packaging complexity.
  - Commercial PDF SDKs: simplify some implementation areas but add license/procurement overhead.

## Decision 2: Watermark suppression strategy

- Decision: Use a batch-informed strategy: estimate shared watermark characteristics across pages, then perform per-page best-effort suppression constrained by readability preservation.
- Rationale: The watermark is repeated and similarly oriented across pages, so cross-page signal aggregation improves robustness and reduces accidental removal of non-watermark content.
- Alternatives considered:
  - Single-page thresholding only: simple but brittle when watermark overlaps text.
  - Frequency-only filtering: may leave artifacts and harm foreground content.
  - ML-based removal: higher setup and model-risk for this scope.

## Decision 3: Image-processing stack

- Decision: Prefer pure-Go image processing (`image`, `x/image`, `imaging`) for grayscale, binarization, and cleanup operations.
- Rationale: Produces a portable, maintainable CLI with minimal runtime dependencies and straightforward CI behavior.
- Alternatives considered:
  - OpenCV via `gocv`: stronger algorithm breadth but requires CGO and native OpenCV distribution.
  - ImageMagick/libvips bindings: capable but dependency-heavy for this feature scope.

## Decision 4: Inconsistent watermark handling policy

- Decision: Continue processing all pages using best-effort suppression instead of failing the entire document.
- Rationale: Matches clarified product expectation and maximizes useful output for long documents where a small subset of pages deviates.
- Alternatives considered:
  - Fail-fast on inconsistency: safer but too restrictive for real-world mixed scans.
  - Skip low-confidence pages: reduces risk but creates incomplete behavior and unpredictable output quality.

## Decision 5: Output naming collision behavior

- Decision: Never overwrite existing outputs; auto-increment output name (e.g., `_output_1.pdf`, `_output_2.pdf`).
- Rationale: Prevents data loss and supports repeatable CLI runs.
- Alternatives considered:
  - Always overwrite existing output.
  - Exit with error on existing output.

## Decision 6: Performance validation target

- Decision: Use a measurable acceptance target of <=5 minutes for typical eligible documents around 100 pages.
- Rationale: Aligns with clarified requirement and provides a concrete benchmark for integration testing.
- Alternatives considered:
  - <=10 minutes: easier to satisfy but weaker user value.
  - <=20 minutes: insufficient for practical batch workflows.
