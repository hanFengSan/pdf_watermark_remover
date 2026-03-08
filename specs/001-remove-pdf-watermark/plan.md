# Implementation Plan: PDF Watermark Removal CLI

**Branch**: `001-remove-pdf-watermark` | **Date**: 2026-03-07 | **Spec**: `/Users/alex/Desktop/works/js/pdf_watermarker_remover/specs/001-remove-pdf-watermark/spec.md`
**Input**: Feature specification from `/Users/alex/Desktop/works/js/pdf_watermarker_remover/specs/001-remove-pdf-watermark/spec.md`

## Summary

Build an offline Go CLI that accepts one image-based PDF, estimates repeated watermark characteristics across pages, suppresses watermark visibility on every page (including best-effort behavior for inconsistent pages), outputs binarized readable pages, and writes a new PDF with `_output` auto-increment naming when needed.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Go standard library; `pdfcpu` for PDF assembly/validation; pure-Go image stack (`image`, `x/image`, `imaging`) for grayscale/binarization/filtering; external renderer command `pdftoppm` (Poppler) for PDF-to-image conversion  
**Storage**: Local filesystem for input/output and temporary page images  
**Testing**: `go test` unit + integration tests; CLI end-to-end verification against `test/test.pdf`  
**Target Platform**: Offline local CLI on macOS/Linux  
**Project Type**: Single-binary CLI utility  
**Performance Goals**: Process eligible ~100-page PDFs in <=5 minutes  
**Constraints**: No overwrite of input, no network requirement, auto-increment output naming, best-effort when watermark style varies  
**Scale/Scope**: One PDF per invocation; minimum 2 pages; typical around 100 pages

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The constitution file at `/Users/alex/Desktop/works/js/pdf_watermarker_remover/.specify/memory/constitution.md` currently contains template placeholders and no ratified enforceable principles.

- Gate A: Defined governance constraints available: PASS (none defined)
- Gate B: Planned approach conflicts with constitution: PASS
- Gate C (post-design re-check): PASS

Post-design re-check result: PASS (no new conflicts introduced by research/design artifacts).

## Project Structure

### Documentation (this feature)

```text
/Users/alex/Desktop/works/js/pdf_watermarker_remover/specs/001-remove-pdf-watermark/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── cli-contract.md
└── tasks.md
```

### Source Code (repository root)

```text
/Users/alex/Desktop/works/js/pdf_watermarker_remover/
├── cmd/
│   └── pdf_watermark_remover/
│       └── main.go
├── internal/
│   ├── pipeline/
│   ├── pdf/
│   ├── watermark/
│   ├── imageproc/
│   └── output/
├── test/
│   └── test.pdf
└── tests/
    ├── integration/
    └── unit/
```

**Structure Decision**: Use a single-project CLI layout with a thin command entrypoint and focused internal domain packages. This supports isolated testing for PDF IO, watermark analysis, and page transforms while keeping user interaction simple.

## Complexity Tracking

No constitution violations require justification.
