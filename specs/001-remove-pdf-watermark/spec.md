# Feature Specification: PDF Watermark Removal

**Feature Branch**: `001-remove-pdf-watermark`  
**Created**: 2026-03-07  
**Status**: Draft  
**Input**: User description: "# 目标 我期望编写一个支持移除PDF水印的软件..."

## Clarifications

### Session 2026-03-07

- Q: 当输出文件已存在时，系统应如何处理？ → A: 自动递增命名（如 `xxx_output_1.pdf`）。
- Q: 当检测到页面间水印样式不一致时，系统应如何处理？ → A: 仍处理全部页面，按最佳猜测去除。
- Q: 对约100页文档，期望处理时长目标是多少？ → A: 5分钟内完成。

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - Remove repeated watermark from scanned PDF (Priority: P1)

As a user, I can provide an image-based PDF with a repeated light, diagonal text watermark across all pages and receive a new PDF where the watermark is no longer visible, so I can use the document without interference.

**Why this priority**: This is the core business outcome; without watermark removal, the feature has no value.

**Independent Test**: Provide a multi-page input PDF containing the same watermark style on every page, run one command, and verify the output PDF has the same page count and no visible watermark text.

**Acceptance Scenarios**:

1. **Given** an input PDF with at least 2 pages and a consistent light diagonal watermark on each page, **When** the user runs the remover command with that file, **Then** the system creates a new PDF named with the `_output.pdf` suffix and watermark text is not visible in the output pages.
2. **Given** an input PDF where page visuals are image-dominant and watermark style is identical across pages, **When** processing completes, **Then** all pages are included in the output and watermark removal behavior is applied consistently across pages.

---

### User Story 2 - Preserve readable document content with binary output (Priority: P2)

As a user, I can accept black-and-white output so that watermark removal can prioritize readability over color fidelity.

**Why this priority**: The user explicitly allows binary output; preserving readability while removing watermark is more important than keeping original colors.

**Independent Test**: Process a sample file and verify output pages are binarized while primary document text and structures remain legible.

**Acceptance Scenarios**:

1. **Given** a valid image-based input PDF, **When** output is generated, **Then** page imagery may be converted to binary style and main document content remains readable for normal use.

---

### User Story 3 - Handle long documents in one run (Priority: P3)

As a user processing large documents (around 100 pages), I can run one command and get one completed output file without page-by-page manual steps.

**Why this priority**: Batch usability is important for practical adoption, but depends on the core watermark-removal flow.

**Independent Test**: Run the feature on a long PDF and confirm it completes end-to-end in one execution with all pages present in output.

**Acceptance Scenarios**:

1. **Given** a valid input PDF with about 100 pages, **When** the user runs the command once, **Then** the system finishes processing and produces one output PDF containing all pages in original order.

---

### Edge Cases

- Input file path does not exist, is unreadable, or is not a PDF.
- Input PDF has only 1 page.
- Input PDF pages do not share a consistent watermark style across all pages (system still processes all pages using best-effort watermark suppression).
- Input PDF has mixed page sizes or orientations.
- Output target filename already exists (system must create an auto-incremented output name instead of overwriting).
- Processing fails on one or more pages in a long document.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST accept a single input PDF file path via command-line invocation.
- **FR-002**: The system MUST validate that the input exists, is readable, and is a PDF before starting processing.
- **FR-003**: The system MUST reject inputs with fewer than 2 pages and return a clear failure message.
- **FR-004**: The system MUST process all pages of a valid input PDF and preserve page order in the output.
- **FR-005**: The system MUST produce an output PDF in the same directory as input using `<input_basename>_output.pdf`; if that name already exists, it MUST create an auto-incremented filename (for example, `<input_basename>_output_1.pdf`).
- **FR-006**: The system MUST identify watermark patterns from page imagery and apply watermark suppression to every page, including best-effort suppression when cross-page watermark consistency is low.
- **FR-007**: The system MUST remove or suppress visible watermark text so that the watermark is no longer readable in normal viewing conditions.
- **FR-008**: The system MUST allow binary (black-and-white) page output and does not need to preserve original colors.
- **FR-009**: The system MUST keep primary non-watermark document content readable after watermark removal.
- **FR-010**: The system MUST return a non-zero exit status and an actionable error message if processing cannot complete.
- **FR-011**: The system MUST complete processing of the provided `test/test.pdf` and generate a corresponding output PDF that satisfies watermark-removal requirements.
- **FR-012**: The system MUST not modify or overwrite the original input PDF.
- **FR-013**: The system MUST complete output generation even when watermark characteristics vary across pages, unless a non-recoverable processing error occurs.
- **FR-014**: For typical eligible documents of around 100 pages, the system MUST target completion within 5 minutes under normal runtime conditions.

### Key Entities

- **Input Document**: A user-provided PDF file; key attributes include file path, page count, readability status, and source filename.
- **Page Image**: The per-page visual representation used for watermark analysis and transformation; key attributes include page index, orientation, and transformed image state.
- **Watermark Pattern**: The repeated visual watermark signature inferred from multiple pages; key attributes include position tendency, angle tendency, intensity tendency, and confidence score.
- **Output Document**: The generated watermark-suppressed PDF; key attributes include output path, page count, generation status, and completion timestamp.

### Assumptions

- Input PDFs are primarily image-based rather than text-layer documents.
- Watermark text appears in a light tone and roughly diagonal (around 45 degrees).
- Watermark wording and style are consistent across all or nearly all pages.
- Users run the command in an environment with enough disk space for temporary processing artifacts.
- For this feature, "watermark removed" means no longer visible/readable during normal viewing, not pixel-perfect restoration of original background.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For eligible inputs (>=2 pages with consistent watermark), 100% of processing runs produce a non-empty output PDF file named with `_output.pdf` or an auto-incremented `_output_<n>.pdf` suffix when needed.
- **SC-002**: On the provided `test/test.pdf`, watermark text is no longer readable on every output page under standard zoom (100%-150%).
- **SC-003**: Output page count matches input page count in 100% of successful runs.
- **SC-004**: At least 95% of pages in a successfully processed document retain readable primary content after watermark suppression.
- **SC-005**: For invalid input conditions (missing file, non-PDF, <2 pages), users receive a clear error result and no output PDF is produced.
- **SC-006**: For eligible documents of approximately 100 pages, end-to-end processing completes within 5 minutes in normal operating conditions.
