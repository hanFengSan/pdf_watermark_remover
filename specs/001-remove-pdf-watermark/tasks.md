# Tasks: PDF Watermark Removal CLI

**Input**: Design documents from `/Users/alex/Desktop/works/js/pdf_watermarker_remover/specs/001-remove-pdf-watermark/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Include tests because the feature explicitly requires processing and verifying `test/test.pdf` and defines measurable acceptance outcomes.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize Go CLI project scaffold and test layout.

- [X] T001 Initialize Go module and baseline dependencies in go.mod
- [X] T002 Create CLI entrypoint scaffold in cmd/pdf_watermark_remover/main.go
- [X] T003 [P] Create pipeline package scaffold in internal/pipeline/runner.go
- [X] T004 [P] Create test package scaffolding in tests/integration/.gitkeep

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core building blocks required by all user stories.

**⚠️ CRITICAL**: No user story work begins until this phase is complete.

- [X] T005 Implement input file validation rules in internal/pipeline/input_validator.go
- [X] T006 [P] Implement output auto-increment naming resolver in internal/output/namer.go
- [X] T007 [P] Implement PDF rasterization wrapper for `pdftoppm` in internal/pdf/render.go
- [X] T008 [P] Implement image-to-PDF rebuild utility in internal/pdf/rebuild.go
- [X] T009 [P] Implement shared processing error types and exit mapping in internal/pipeline/errors.go
- [X] T010 Implement end-to-end pipeline orchestration skeleton in internal/pipeline/runner.go
- [X] T011 Add integration test helpers for fixture and output cleanup in tests/integration/helpers_test.go

**Checkpoint**: Foundation ready - user story implementation can begin.

---

## Phase 3: User Story 1 - Remove repeated watermark from scanned PDF (Priority: P1) 🎯 MVP

**Goal**: Accept a multi-page image-based PDF and produce an output PDF where repeated watermark text is no longer readable.

**Independent Test**: Run `pdf_watermark_remover test/test.pdf`, then verify output exists, page count matches input, and watermark text is not readable at 100%-150% zoom.

### Tests for User Story 1

- [X] T012 [P] [US1] Add CLI contract test for required input argument and exit codes in tests/contract/cli_contract_test.go
- [X] T013 [P] [US1] Add integration test for watermark suppression on test fixture in tests/integration/us1_watermark_removal_test.go

### Implementation for User Story 1

- [X] T014 [P] [US1] Implement cross-page watermark pattern estimator in internal/watermark/pattern_estimator.go
- [X] T015 [P] [US1] Implement per-page watermark suppression transform in internal/watermark/suppressor.go
- [X] T016 [US1] Integrate pattern estimation and suppression into processing flow in internal/pipeline/runner.go
- [X] T017 [US1] Implement CLI execution flow and success output path messaging in cmd/pdf_watermark_remover/main.go
- [X] T018 [US1] Enforce non-recoverable failure behavior with non-zero exit in internal/pipeline/errors.go
- [X] T019 [US1] Enforce original input non-modification guarantees in internal/pipeline/input_validator.go

**Checkpoint**: User Story 1 is independently functional and testable.

---

## Phase 4: User Story 2 - Preserve readable document content with binary output (Priority: P2)

**Goal**: Convert pages to binary output while preserving readability of primary non-watermark content.

**Independent Test**: Process `test/test.pdf` and verify output pages are binarized while primary text/content remains readable.

### Tests for User Story 2

- [X] T020 [P] [US2] Add integration test for binarized output readability in tests/integration/us2_binarized_readability_test.go
- [X] T021 [P] [US2] Add unit tests for thresholding behavior in tests/unit/binarize_test.go

### Implementation for User Story 2

- [X] T022 [P] [US2] Implement grayscale and adaptive binarization stage in internal/imageproc/binarize.go
- [X] T023 [P] [US2] Implement readability-preservation guard heuristics in internal/imageproc/readability_guard.go
- [X] T024 [US2] Wire binarization and readability guard into pipeline stages in internal/pipeline/runner.go
- [X] T025 [US2] Document binary-output behavior in CLI help text in cmd/pdf_watermark_remover/main.go

**Checkpoint**: User Stories 1 and 2 both work independently.

---

## Phase 5: User Story 3 - Handle long documents in one run (Priority: P3)

**Goal**: Process long documents (~100 pages) in one run with best-effort behavior for inconsistent watermark pages and target runtime <=5 minutes.

**Independent Test**: Run on a long eligible PDF and verify all pages are output in order, inconsistent pages are processed best-effort, and run time meets target in normal conditions.

### Tests for User Story 3

- [X] T026 [P] [US3] Add integration test for full-document page-order preservation in tests/integration/us3_batch_processing_test.go
- [X] T027 [P] [US3] Add integration test for best-effort mode on watermark-inconsistent pages in tests/integration/us3_best_effort_test.go
- [X] T028 [P] [US3] Add integration performance assertion test for <=5 minute target in tests/integration/us3_performance_target_test.go

### Implementation for User Story 3

- [X] T029 [P] [US3] Implement worker-pool page processing for throughput in internal/pipeline/worker_pool.go
- [X] T030 [P] [US3] Implement page-level best-effort mode selector in internal/watermark/mode_selector.go
- [X] T031 [US3] Integrate worker pool and best-effort fallback in internal/pipeline/runner.go
- [X] T032 [US3] Implement processing progress and duration reporting in internal/pipeline/progress.go
- [X] T033 [US3] Expose final processing summary in CLI output in cmd/pdf_watermark_remover/main.go

**Checkpoint**: All user stories are independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Hardening and final verification across stories.

- [X] T034 [P] Add unit tests for output name collision edge cases in tests/unit/output_namer_test.go
- [X] T035 [P] Add unit tests for input validation edge cases in tests/unit/input_validator_test.go
- [X] T036 Update quickstart verification checklist with final commands in specs/001-remove-pdf-watermark/quickstart.md
- [X] T037 Execute full test suite and record expected validation flow in specs/001-remove-pdf-watermark/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies.
- **Phase 2 (Foundational)**: Depends on Phase 1; blocks all user stories.
- **Phase 3 (US1)**: Depends on Phase 2; defines MVP.
- **Phase 4 (US2)**: Depends on Phase 2 and reuses US1 pipeline hooks.
- **Phase 5 (US3)**: Depends on Phase 2 and reuses US1 core flow.
- **Phase 6 (Polish)**: Depends on completion of desired user stories.

### User Story Dependencies

- **US1 (P1)**: Independent after foundational completion; no dependency on other stories.
- **US2 (P2)**: Independent acceptance criteria, but implementation extends shared processing path introduced in US1.
- **US3 (P3)**: Independent acceptance criteria, but implementation extends shared processing path introduced in US1.

### Within Each User Story

- Write tests first and confirm they fail before implementation.
- Implement core processing components before pipeline integration.
- Complete story-specific validation before moving to next story.

### Parallel Opportunities

- Phase 1: T003 and T004 can run in parallel.
- Phase 2: T006, T007, T008, and T009 can run in parallel after T005 starts.
- US1: T012/T013 and T014/T015 pairs can run in parallel.
- US2: T020/T021 and T022/T023 pairs can run in parallel.
- US3: T026/T027/T028 and T029/T030 can run in parallel.
- Polish: T034 and T035 can run in parallel.

---

## Parallel Example: User Story 1

```bash
# Parallel tests
Task: "Add CLI contract test for required input argument and exit codes in tests/contract/cli_contract_test.go"
Task: "Add integration test for watermark suppression on test fixture in tests/integration/us1_watermark_removal_test.go"

# Parallel implementation components
Task: "Implement cross-page watermark pattern estimator in internal/watermark/pattern_estimator.go"
Task: "Implement per-page watermark suppression transform in internal/watermark/suppressor.go"
```

## Parallel Example: User Story 2

```bash
# Parallel tests
Task: "Add integration test for binarized output readability in tests/integration/us2_binarized_readability_test.go"
Task: "Add unit tests for thresholding behavior in tests/unit/binarize_test.go"

# Parallel implementation components
Task: "Implement grayscale and adaptive binarization stage in internal/imageproc/binarize.go"
Task: "Implement readability-preservation guard heuristics in internal/imageproc/readability_guard.go"
```

## Parallel Example: User Story 3

```bash
# Parallel tests
Task: "Add integration test for full-document page-order preservation in tests/integration/us3_batch_processing_test.go"
Task: "Add integration test for best-effort mode on watermark-inconsistent pages in tests/integration/us3_best_effort_test.go"
Task: "Add integration performance assertion test for <=5 minute target in tests/integration/us3_performance_target_test.go"

# Parallel implementation components
Task: "Implement worker-pool page processing for throughput in internal/pipeline/worker_pool.go"
Task: "Implement page-level best-effort mode selector in internal/watermark/mode_selector.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 (Setup).
2. Complete Phase 2 (Foundational).
3. Complete Phase 3 (US1).
4. Validate by running the US1 independent test using `test/test.pdf`.

### Incremental Delivery

1. Deliver US1 (core watermark removal).
2. Deliver US2 (binary readability improvements).
3. Deliver US3 (long-document throughput and best-effort behavior).
4. Finish with polish tasks and full-suite validation.

### Parallel Team Strategy

1. One developer completes Phase 1 and Phase 2 core orchestration tasks.
2. After foundation: split US1/US2/US3 implementation components marked [P].
3. Merge each story only after story-specific tests pass independently.

---

## Notes

- [P] tasks are safe parallel work items on different files.
- [USx] labels map tasks back to user stories for independent delivery.
- Keep each story releasable and testable without requiring completion of later stories.
